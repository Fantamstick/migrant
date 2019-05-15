package app

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/spf13/viper"
)

type secretHash map[string]string

var secretSources map[string]secretHash

// SECRET_PROTOCOL defines the string that should be present in config values to signify
// that value is a secret.
const SECRET_PROTOCOL = "SECRET://"

// LoadSecrets loads secretSources from all sources specified in config secrets block.
func LoadSecrets() error {

	// create new map to store secretSources
	secretSources = make(map[string]secretHash)

	// load secret block from config
	block := viper.GetStringMap("secrets")

	for s := range block {
		source := viper.Sub("secrets").GetStringMapString(s)

		// If a source with the specified name already exists, do not allow import.
		// This should not happen, since it would require a collision in the config file.
		if _, alreadyExists := secretSources[s]; alreadyExists {
			return NewErrSecretsAlreadyDefined(s)
		}

		var driver string
		var uri string
		var ok bool

		if driver, ok = source["driver"]; !ok {
			return fmt.Errorf("driver not present in secrets block for source:" + s)
		}

		if uri, ok = source["uri"]; !ok {
			return fmt.Errorf("uri not present in secrets block for source:" + s)
		}

		var err error

		// use the appropriate loader based on the source driver
		switch driver {
		case "aws-secretsmanager":
			err = loadAwsSMSecrets(s, uri)
		case "json":
			err = loadJsonSecrets(s, uri)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// load secrets from aws secret manager. Will try to connect to aws using the credentials in the user's .aws folder.
// if the credentials don't exist or there are insufficient permissions, will return an error. Checks for secrets
// manager errors as per aws example code.
func loadAwsSMSecrets(name, uri string) error {
	URI, err := url.Parse(uri)

	if err != nil {
		return err
	}

	params := URI.Query()
	region := params.Get("region")

	if region == "" {
		return NewErrBadConfig("you must specifiy a region for the aws-secretsmanager source: " + name)
	}

	secretName := URI.EscapedPath()

	config := aws.Config{
		Region: aws.String(region),
	}

	session, err := session.NewSession(&config)

	if err != nil {
		panic(err)
	}

	service := secretsmanager.New(session)

	result, err := service.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {

			case secretsmanager.ErrCodeDecryptionFailure:
				return NewErrAwsException(secretsmanager.ErrCodeDecryptionFailure, aerr.Error())

			case secretsmanager.ErrCodeInternalServiceError:
				return NewErrAwsException(secretsmanager.ErrCodeInternalServiceError, aerr.Error())

			case secretsmanager.ErrCodeInvalidParameterException:
				return NewErrAwsException(secretsmanager.ErrCodeInvalidParameterException, aerr.Error())

			case secretsmanager.ErrCodeInvalidRequestException:
				return NewErrAwsException(secretsmanager.ErrCodeInvalidRequestException, aerr.Error())

			case secretsmanager.ErrCodeResourceNotFoundException:
				return NewErrAwsException(secretsmanager.ErrCodeResourceNotFoundException, aerr.Error())

			case "AccessDeniedException":
				return NewErrAwsException("AccessDeniedException", aerr.Error())
			}

		} else {
			return NewErrAwsException("Unknown", err.Error())
		}
	}

	// get secret string (may ether be a string or encoded binary blob)
	var secret string

	if result.SecretString != nil {
		secret = *result.SecretString
	} else {
		bytes := make([]byte, base64.StdEncoding.DecodedLen(len(result.SecretBinary)))
		l, err := base64.StdEncoding.Decode(bytes, result.SecretBinary)
		if err != nil {
			return err
		}
		secret = string(bytes[:l])
	}

	// parse secret string into json
	var secHash secretHash
	err = json.Unmarshal([]byte(secret), &secHash)

	if err != nil {
		panic(err)
	}

	// store secrets
	secretSources[name] = secHash

	return nil
}

// load secrets from a json file. This method is intended for facilitating internal testing, and not actual production
// use - however, if you put your secrets into a well gaurded file, then it would at least make a workable solution for
// storing local secrets.
func loadJsonSecrets(name, uri string) error {
	info, err := os.Stat(uri)

	if err != nil {
		return err
	}

	if info.IsDir() {
		return NewErrBadConfig("secrets block uri must specify a json file, not a path: " + uri)
	}

	var bytes []byte
	bytes, err = ioutil.ReadFile(uri)

	if err != nil {
		return err
	}

	var sh secretHash
	err = json.Unmarshal(bytes, &sh)

	if err != nil {
		return err
	}

	secretSources[name] = sh

	return nil
}

// MustLoadSecrets loads secrets or panics.
func MustLoadSecrets() {
	err := LoadSecrets()
	if err != nil {
		panic(err)
	}
}

// Secret returns a single secret. Secrets must be loaded into memory before this function is called.
// if a string is passed that is not a secret, the string is returned unchanged with no errors
func Secret(uri string) (string, error) {
	if !IsSecretUri(uri) {
		return uri, nil
	}

	if secretSources == nil {
		return "", fmt.Errorf("must load secrets before accessing them")
	}

	if strings.Contains(uri, SECRET_PROTOCOL) {
		uri = strings.TrimLeft(uri, SECRET_PROTOCOL)
	}

	set := strings.Split(uri, "/")

	if len(set) < 2 {
		return "", NewErrBadConfig("secret uri must be comprised of two parts - check your secrets block")
	}

	sourceName := set[0]
	secretName := set[1]

	var source secretHash
	var ok bool

	if source, ok = secretSources[sourceName]; !ok {
		return "", NewErrSecretNotFound(sourceName, secretName)
	}

	var secret string

	if secret, ok = source[secretName]; !ok {
		return "", NewErrSecretNotFound(sourceName, secretName)
	}

	return secret, nil
}

// NeedSecret will return the specified secret uri or panic if anything fails.
func NeedSecret(uri string) string {
	s, err := Secret(uri)

	if err != nil {
		panic("error getting secret: " + err.Error())
	}

	return s
}

// IsSecretUri returns true if the passed string is *trivially* identifiable as a secret uri.
func IsSecretUri(uri string) bool {
	return strings.Contains(uri, SECRET_PROTOCOL)
}
