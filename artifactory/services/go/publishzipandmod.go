package _go

import (
	"net/url"
	"strings"

	"github.com/jfrog/jfrog-client-go/artifactory/auth"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/errors/httperrors"
	"github.com/jfrog/jfrog-client-go/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/version"
)

func init() {
	register(&publishZipAndModApi{})
}

// Support for Artifactory 6.6.1 and above API
type publishZipAndModApi struct {
}

func (pwa *publishZipAndModApi) isCompatible(artifactoryVersion string) bool {
	propertiesApi := "6.6.1"
	if version.Compare(artifactoryVersion, propertiesApi) < 0 && artifactoryVersion != "development" {
		return false
	}
	return true
}

func (pwa *publishZipAndModApi) PublishPackage(params GoParams, client *httpclient.HttpClient, ArtDetails auth.ArtifactoryDetails) error {
	url, err := utils.BuildArtifactoryUrl(ArtDetails.GetUrl(), "api/go/"+params.GetTargetRepo(), make(map[string]string))
	if err != nil {
		return err
	}
	zipUrl := url
	moduleId := strings.Split(params.GetModuleId(), ":")
	err = createUrlPath(moduleId[0], params.GetVersion(), params.GetProps(), ".zip", &zipUrl)
	if err != nil {
		return err
	}
	clientDetails := ArtDetails.CreateHttpClientDetails()

	addGoVersion(params, &zipUrl)
	resp, body, err := client.UploadFileWithTimeoutRetry(params.GetZipPath(), zipUrl, clientDetails, 2, 10)
	if err != nil {
		return err
	}

	// Forbiden error might be received by a reattempt when the previous attempt
	// times out even though it has been processed by the server, and the user
	// does not have ovewrite permission. That is why we are expecting that error
	// status here and proceeding with the mod file upload
	err = httperrors.CheckResponseStatus(resp, body, 201, 403)
	if err != nil {
		return err
	}
	err = createUrlPath(moduleId[0], params.GetVersion(), params.GetProps(), ".mod", &url)
	if err != nil {
		return err
	}
	addGoVersion(params, &url)
	resp, body, err = client.UploadFileWithTimeoutRetry(params.GetModPath(), url, clientDetails, 2, 10)
	if err != nil {
		return err
	}
	return httperrors.CheckResponseStatus(resp, body, 201)
}

func addGoVersion(params GoParams, urlPath *string) {
	*urlPath += ";go.version=" + url.QueryEscape(params.GetVersion())
}
