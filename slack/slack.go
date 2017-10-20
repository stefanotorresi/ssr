package slack

import (
	"encoding/json"
	"net/http"
	"net/url"
)

const baseUrl = "https://Api.slack.com/Api/"

type Api struct {
	authToken string
	client    *http.Client
}

type ApiError struct {
	msg string
}

type Response struct {
	Ok            bool   `json:"ok"`
	Error         string `json:"error"`
	RequestMethod string `json:"req_method"`
}

type ChannelMembers struct {
	Response
	Members []string `json:"members"`
}

type UserInfo struct {
	Response
	User struct {
		Id      string `json:"id"`
		Profile struct {
			DisplayName        string `json:"display_name"`
			RealName           string `json:"real_name"`
			RealNameNormalized string `json:"real_name_normalized"`
			Email              string `json:"email"`
		} `json:"profile"`
	} `json:"user"`
}

func New(authToken string, client *http.Client) *Api {
	return &Api{authToken, client}
}

func (e *ApiError) Error() string {
	return "API ApiError: " + e.msg
}

func isHttpError(r *http.Response) bool {
	return r.StatusCode >= 400 && r.StatusCode < 600
}

func (api *Api) FetchChannelMembers(channelId string) (ChannelMembers, error) {
	endpoint := "conversations.members"

	response, err := api.client.PostForm(baseUrl+endpoint, url.Values{
		"token":   {api.authToken},
		"channel": {channelId},
	})
	if err != nil {
		return ChannelMembers{}, &ApiError{err.Error()}
	}
	if isHttpError(response) {
		return ChannelMembers{}, &ApiError{response.Status}
	}

	defer response.Body.Close()

	return decodeChannelMembers(response)
}

func decodeChannelMembers(response *http.Response) (ChannelMembers, error) {
	channelMembers := ChannelMembers{}
	err := json.NewDecoder(response.Body).Decode(&channelMembers)

	if err == nil && !channelMembers.Ok {
		err = &ApiError{channelMembers.Error}
	}

	return channelMembers, err
}
