package responsehandler

import (
	"github.com/Fabriciope/cli_chat/client/cui"
	"github.com/Fabriciope/cli_chat/pkg/escapecode"
	"github.com/Fabriciope/cli_chat/pkg/shared/dto"
)

func (handler *ResponseHandler) Login(response dto.Response) {
	payload := response.Payload.(string)

	if response.Err {
		handler.cui.DrawLoginInterfaceWithMessage(payload, escapecode.Red)
		return
	}

	*handler.userLoggedIn = true
	handler.cui.RenderChatInterface()
	handler.cui.PrintLine(&cui.Line{
		Info:      "login status:",
		Text:      response.Payload.(string),
		TextColor: escapecode.Green,
	})
}
