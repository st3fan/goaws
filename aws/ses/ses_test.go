package ses

import (
	"log"
	"testing"
)

func Test_parseListIdentitiesResponse(t *testing.T) {
	xml := `<ListIdentitiesResponse xmlns="http://ses.amazonaws.com/doc/2010-12-01/">
            <ListIdentitiesResult>
              <Identities>
                <member>example.com</member>
                <member>user@example.com</member>
              </Identities>
            </ListIdentitiesResult>
            <ResponseMetadata>
              <RequestId>cacecf23-9bf1-11e1-9279-0100e8cf109a</RequestId>
            </ResponseMetadata>
          </ListIdentitiesResponse>`
	response, err := parseListIdentitiesResponse([]byte(xml))
	if err != nil {
		t.Error("parseListIdentitiesResponse failed", err)
	}

	if response.ResponseMetadata.RequestId != "cacecf23-9bf1-11e1-9279-0100e8cf109a" {
		t.Error("ListIdentitiesResult.ResponseMetadata.RequestId failed to parse")
	}

	if len(response.ListIdentitiesResult.Identities) != 2 {
		t.Error("ListIdentitiesResult.Identities failed to parse")
	}

	if response.ListIdentitiesResult.Identities[0] != "example.com" {
		t.Error("ListIdentitiesResult.Identities[0] failed to parse")
	}

	if response.ListIdentitiesResult.Identities[1] != "user@example.com" {
		t.Error("ListIdentitiesResult.Identities[1] failed to parse")
	}
}

func Test_parseErrorResponse(t *testing.T) {
	xml := `<ErrorResponse>
   <Error>
      <Type>
         Sender
      </Type>
      <Code>
         ValidationError
      </Code>
      <Message>
         Value null at 'message.subject' failed to satisfy constraint: Member must not be null
      </Message>
   </Error>
   <RequestId>
      42d59b56-7407-4c4a-be0f-4c88daeea257
   </RequestId>
</ErrorResponse>`
	response, err := parseErrorResponse([]byte(xml))
	if err != nil {
		t.Error("parseErrorResponse failed", err)
	}

	log.Printf("%+v\n", response)

	if response.RequestId != "42d59b56-7407-4c4a-be0f-4c88daeea257" {
		t.Error("response.RequestId failed to parse")
	}
}
