package ses

import (
	"encoding/xml"
	"fmt"
	"github.com/st3fan/mailer/aws"
	"log"
)

// This should all move to an aws/ses package

type IdentityType string

const (
	EmailAddressIdentityType IdentityType = "EmailAddress"
	DomainIdentityType                    = "Domain"
	AnyIdentityType                       = "Any"
)

type SimpleEmailService struct {
	credentials aws.Credentials
	endpoint    string
}

type Destination struct {
	BccAddresses []string
	CcAddresses  []string
	ToAddresses  []string
}

type Content struct {
	Charset string
	Data    string
}

type Body struct {
	Html Content
	Text Content
}

type Message struct {
	Body    Body
	Subject Content
}

type VerifyEmailIdentityResponse struct {
	ResponseMetadata aws.ResponseMetadata
}

type DeleteIdentityResponse struct {
	ResponseMetadata aws.ResponseMetadata
}

type ListIdentitiesResponse struct {
	ListIdentitiesResult struct {
		Identities []string `xml:"Identities>member"`
	}
	ResponseMetadata aws.ResponseMetadata
}

type IdentityVerificationAttributes struct {
	VerificationStatus string
	VerificationToken  string
}

type RawGetIdentityVerificationAttributesResponse struct {
	GetIdentityVerificationAttributesResult struct {
		VerificationAttributes struct {
			Entries []struct {
				Key   string                         `xml:"key"`
				Value IdentityVerificationAttributes `xml:"value"`
			} `xml:"entry"`
		}
	}
	ResponseMetadata aws.ResponseMetadata
}

type GetIdentityVerificationAttributesResponse struct {
	GetIdentityVerificationAttributesResult struct {
		VerificationAttributes map[string]IdentityVerificationAttributes
	}
	ResponseMetadata aws.ResponseMetadata
}

type SendEmailResponse struct {
	SendEmailResult struct {
		MessageId string
	}
	ResponseMetadata aws.ResponseMetadata
}

func endpointForRegion(region aws.Region) (string, error) {
	switch region {
	case aws.USEast1Region:
		return "https://email.us-east-1.amazonaws.com", nil
	case aws.USWest2Region:
		return "https://email.us-west-2.amazonaws.com", nil
	case aws.EUWest1Region:
		return "https://email.eu-west-1.amazonaws.com", nil
	}
	return "", aws.UnsupportedRegionErr
}

func NewSimpleEmailService(credentials aws.Credentials, region aws.Region) (*SimpleEmailService, error) {
	endpoint, err := endpointForRegion(region)
	if err != nil {
		return nil, err
	}
	return &SimpleEmailService{credentials: credentials, endpoint: endpoint}, nil
}

func parseVerifyEmailIdentityResponse(data []byte) (VerifyEmailIdentityResponse, error) {
	response := VerifyEmailIdentityResponse{}
	if err := xml.Unmarshal(data, &response); err != nil {
		return VerifyEmailIdentityResponse{}, err
	}
	return response, nil
}

func (ses *SimpleEmailService) VerifyEmailIdentity(emailAddress string) (VerifyEmailIdentityResponse, error) {
	parameters := aws.Parameters{"EmailAddress": emailAddress}
	data, err := aws.ExecuteRequest(ses.credentials, ses.endpoint, "/", "VerifyEmailIdentity", parameters)
	if err != nil {
		return VerifyEmailIdentityResponse{}, err
	}
	return parseVerifyEmailIdentityResponse(data)
}

func parseListIdentitiesResponse(data []byte) (ListIdentitiesResponse, error) {
	listIdentitiesResponse := ListIdentitiesResponse{}
	if err := xml.Unmarshal(data, &listIdentitiesResponse); err != nil {
		return ListIdentitiesResponse{}, err
	}
	return listIdentitiesResponse, nil
}

func (ses *SimpleEmailService) ListIdentities(identityType IdentityType, maxItems int, nextToken string) (ListIdentitiesResponse, error) {
	parameters := aws.Parameters{}
	if identityType != AnyIdentityType {
		parameters["IdentityType"] = identityType
	}
	if maxItems != 0 {
		parameters["MaxItems"] = maxItems
	}
	if nextToken != "" {
		parameters["NextToken"] = nextToken
	}
	data, err := aws.ExecuteRequest(ses.credentials, ses.endpoint, "/", "ListIdentities", parameters)
	if err != nil {
		return ListIdentitiesResponse{}, err
	}
	return parseListIdentitiesResponse(data)
}

func parseDeleteIdentityResponse(data []byte) (DeleteIdentityResponse, error) {
	response := DeleteIdentityResponse{}
	if err := xml.Unmarshal(data, &response); err != nil {
		return DeleteIdentityResponse{}, err
	}
	return response, nil
}

func (ses *SimpleEmailService) DeleteIdentity(identity string) (DeleteIdentityResponse, error) {
	parameters := aws.Parameters{"Identity": identity}
	data, err := aws.ExecuteRequest(ses.credentials, ses.endpoint, "/", "DeleteIdentity", parameters)
	if err != nil {
		return DeleteIdentityResponse{}, err
	}
	return parseDeleteIdentityResponse(data)
}

func parseGetIdentityVerificationAttributesResponse(data []byte) (GetIdentityVerificationAttributesResponse, error) {
	rawResponse := RawGetIdentityVerificationAttributesResponse{}
	if err := xml.Unmarshal(data, &rawResponse); err != nil {
		return GetIdentityVerificationAttributesResponse{}, err
	}

	log.Printf("%v\n", rawResponse)

	response := GetIdentityVerificationAttributesResponse{}
	response.GetIdentityVerificationAttributesResult.VerificationAttributes = make(map[string]IdentityVerificationAttributes)
	for _, v := range rawResponse.GetIdentityVerificationAttributesResult.VerificationAttributes.Entries {
		response.GetIdentityVerificationAttributesResult.VerificationAttributes[v.Key] = v.Value
	}
	response.ResponseMetadata = rawResponse.ResponseMetadata

	return response, nil
}

func (ses *SimpleEmailService) GetIdentityVerificationAttributes(identities []string) (GetIdentityVerificationAttributesResponse, error) {
	parameters := aws.Parameters{"Identities.member": identities}
	data, err := aws.ExecuteRequest(ses.credentials, ses.endpoint, "/", "GetIdentityVerificationAttributes", parameters)
	if err != nil {
		return GetIdentityVerificationAttributesResponse{}, err
	}
	return parseGetIdentityVerificationAttributesResponse(data)
}

//

func parseSendEmailResponse(data []byte) (SendEmailResponse, error) {
	response := SendEmailResponse{}
	if err := xml.Unmarshal(data, &response); err != nil {
		return SendEmailResponse{}, err
	}
	return response, nil
}

func (ses *SimpleEmailService) SendEmail(destination Destination, message Message, replyToAddresses []string, returnPath string, source string) (SendEmailResponse, error) {
	parameters := aws.Parameters{}
	if message.Body.Text.Data != "" {
		parameters["Message.Body.Text.Data"] = message.Body.Text.Data
	}
	if message.Body.Text.Charset != "" {
		parameters["Message.Body.Text.Charset"] = message.Body.Text.Charset
	}
	if message.Body.Html.Data != "" {
		parameters["Message.Body.Html.Data"] = message.Body.Html.Data
	}
	if message.Body.Html.Charset != "" {
		parameters["Message.Body.Html.Charset"] = message.Body.Html.Charset
	}
	if message.Subject.Data != "" {
		parameters["Message.Subject.Data"] = message.Subject.Data
	}
	if message.Subject.Charset != "" {
		parameters["Message.Subject.Charset"] = message.Subject.Charset
	}
	for idx, value := range destination.ToAddresses {
		parameters[fmt.Sprintf("Destination.ToAddresses.member.%d", idx+1)] = value
	}
	for idx, value := range destination.CcAddresses {
		parameters[fmt.Sprintf("Destination.CcAddresses.member.%d", idx+1)] = value
	}
	for idx, value := range destination.BccAddresses {
		parameters[fmt.Sprintf("Destination.BccAddresses.member.%d", idx+1)] = value
	}
	for idx, value := range replyToAddresses {
		parameters[fmt.Sprintf("ReplyToAddresses.member.%d", idx+1)] = value
	}
	if returnPath != "" {
		parameters["ReturnPath"] = returnPath
	}
	if source != "" {
		parameters["Source"] = source
	}
	data, err := aws.ExecuteRequest(ses.credentials, ses.endpoint, "/", "SendEmail", parameters)
	if err != nil {
		return SendEmailResponse{}, err
	}
	return parseSendEmailResponse(data)
}
