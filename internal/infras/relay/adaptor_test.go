package relay

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"hermes-ai/internal/infras/relay/apitype"
)

func TestGetAdaptor(t *testing.T) {
	Convey("get adaptor", t, func() {
		for i := 0; i < apitype.Dummy; i++ {
			a := GetAdaptor(i)
			So(a, ShouldNotBeNil)
		}
	})
}
