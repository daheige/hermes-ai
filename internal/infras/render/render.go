package render

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/infras/customevent"
)

func StringData(c *gin.Context, str string) {
	str = strings.TrimPrefix(str, "data: ")
	str = strings.TrimSuffix(str, "\r")
	c.Render(-1, customevent.CustomEvent{Data: "data: " + str})
	c.Writer.Flush()
}

func ObjectData(c *gin.Context, object interface{}) error {
	jsonData, err := json.Marshal(object)
	if err != nil {
		return fmt.Errorf("error marshalling object: %w", err)
	}
	StringData(c, string(jsonData))
	return nil
}

func Done(c *gin.Context) {
	StringData(c, "[DONE]")
}
