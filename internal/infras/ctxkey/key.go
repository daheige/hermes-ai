package ctxkey

// CtxKey context key
type CtxKey struct {
	Name string
}

func (c *CtxKey) String() string {
	return c.Name
}

var (
	Config            = CtxKey{"config"}
	Id                = CtxKey{"id"}
	Username          = CtxKey{"username"}
	Role              = CtxKey{"role"}
	Status            = CtxKey{"status"}
	Channel           = CtxKey{"channel"}
	ChannelId         = CtxKey{"channel_id"}
	SpecificChannelId = CtxKey{"specific_channel_id"}
	RequestModel      = CtxKey{"request_model"}
	ConvertedRequest  = CtxKey{"converted_request"}
	OriginalModel     = CtxKey{"original_model"}
	Group             = CtxKey{"group"}
	ModelMapping      = CtxKey{"model_mapping"}
	ChannelName       = CtxKey{"channel_name"}
	TokenId           = CtxKey{"token_id"}
	TokenName         = CtxKey{"token_name"}
	BaseURL           = CtxKey{"base_url"}
	AvailableModels   = CtxKey{"available_models"}
	KeyRequestBody    = CtxKey{"key_request_body"}
	SystemPrompt      = CtxKey{"system_prompt"}
	RequestIdKey      = CtxKey{"X-Request-Id"}
	ResponseFormat    = CtxKey{"response_format"}
)
