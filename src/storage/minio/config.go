package minio

type MinioConfig struct {
	Url             string
	Port            int
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
}

func LoadMinioConfigFromEnv(theConfig MinioConfig) MinioConfig {

	return MinioConfig{
		Url:             theConfig.Url,
		Port:            theConfig.Port,
		AccessKeyID:     theConfig.AccessKeyID,
		SecretAccessKey: theConfig.SecretAccessKey,
		UseSSL:          theConfig.UseSSL,
	}
}
