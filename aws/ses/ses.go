package ses

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"github.com/st3fan/goaws/aws"
	"log"
	"strings"
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

func NewSingleDestination(to string) Destination {
	return Destination{
		ToAddresses: []string{to},
	}
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

func NewTextMessage(subject, body string) Message {
	return Message{
		Body: Body{
			Text: Content{
				Charset: "UTF-8",
				Data:    body,
			},
		},
		Subject: Content{
			Charset: "UTF-8",
			Data:    subject,
		},
	}
}

type RawMessage struct {
	Data string
}

func NewRawMessage(data string) RawMessage {
	return RawMessage{Data: data}
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

type SendRawEmailResponse struct {
	SendRawEmailResult struct {
		MessageId string
	}
	ResponseMetadata aws.ResponseMetadata
}

type ErrorResponse struct {
	Error struct {
		Type    string
		Code    string
		Message string
	}
	RequestId string
}

func parseErrorResponse(data []byte) (ErrorResponse, error) {
	response := ErrorResponse{}
	if err := xml.Unmarshal(data, &response); err != nil {
		return ErrorResponse{}, err
	}

	// Usually AWS APIs are pretty consistent. This one is not.
	response.Error.Type = strings.TrimSpace(response.Error.Type)
	response.Error.Code = strings.TrimSpace(response.Error.Code)
	response.Error.Message = strings.TrimSpace(response.Error.Message)
	response.RequestId = strings.TrimSpace(response.RequestId)

	return response, nil
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

func parseSendRawEmailResponse(data []byte) (SendRawEmailResponse, error) {
	response := SendRawEmailResponse{}
	if err := xml.Unmarshal(data, &response); err != nil {
		return SendRawEmailResponse{}, err
	}
	return response, nil
}

func (ses *SimpleEmailService) SendRawEmail(destinations []string, rawMessage RawMessage, source string) (SendRawEmailResponse, error) {
	parameters := aws.Parameters{}
	for idx, value := range destinations {
		parameters[fmt.Sprintf("Destinations.member.%d", idx+1)] = value
	}
	parameters["RawMessage.Data"] = base64.StdEncoding.EncodeToString([]byte(rawMessage.Data))
	if source != "" {
		parameters["Source"] = source
	}
	data, err := aws.ExecuteRequest(ses.credentials, ses.endpoint, "/", "SendRawEmail", parameters)
	if err != nil {
		return SendRawEmailResponse{}, err
	}
	return parseSendRawEmailResponse(data)
}
