package e2e

import (
	"encoding/json"
	"encoding/xml"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCases(t *testing.T) {
	var xmlMarshalCases = []struct {
		Name     string
		Value    interface{}
		Expected string
	}{
		{
			"with enum in root",
			&MessageWithRootEnum{
				Field: RootEnum_DEF,
			},
			"<MessageWithRootEnum><Field>DEF</Field></MessageWithRootEnum>",
		},

		{
			"with nested enum",
			&MessageWithNestedEnum{
				Field: Nested_DEF,
			},
			"<MessageWithNestedEnum><Field>DEF</Field></MessageWithNestedEnum>",
		},

		{
			"with deeply nested enum",
			&MessageWithDeeplyNestedEnum{
				Field: Deeply_Nested_DEF,
			},
			"<MessageWithDeeplyNestedEnum><Field>DEF</Field></MessageWithDeeplyNestedEnum>",
		},
	}

	for _, tt := range xmlMarshalCases {
		t.Run(tt.Name, func(t *testing.T) {
			require := require.New(t)

			bs, err := xml.Marshal(tt.Value)

			require.NoError(err)
			require.NotEmpty(bs)
			require.Equal(tt.Expected, string(bs))

			val := reflect.New(reflect.ValueOf(tt.Value).Elem().Type())

			require.NoError(xml.Unmarshal(bs, val.Interface()))
			require.Equal(val.Interface(), tt.Value)
		})
	}

	var xmlUnmarshalCases = []struct {
		Name     string
		Value    string
		Expected interface{}
	}{
		{
			"match snake case with enum prefix",
			"<ScreamingSnakeWithPrefxEnum><Field>DEF</Field></ScreamingSnakeWithPrefxEnum>",
			&ScreamingSnakeWithPrefxEnum{
				Field: ScreamingSnakeWithPrefix_SCREAMING_SNAKE_WITH_PREFIX_DEF,
			},
		},

		{
			"unmarshal attribute",
			"<ScreamingSnakeWithPrefxEnum field=\"DEF\" />",
			&MessageWithAttribute{
				Field: RootEnum_DEF,
			},
		},
	}

	for _, tt := range xmlUnmarshalCases {
		t.Run(tt.Name, func(t *testing.T) {
			require := require.New(t)
			val := reflect.New(reflect.ValueOf(tt.Expected).Elem().Type())

			require.NoError(xml.Unmarshal([]byte(tt.Value), val.Interface()))
			require.Equal(tt.Expected, val.Interface())
		})
	}

	var jsonMarshalCases = []struct {
		Name     string
		Value    interface{}
		Expected string
	}{
		{
			"with enum in root",
			&MessageWithRootEnum{
				Field: RootEnum_DEF,
			},
			"{\"field\":\"DEF\"}",
		},

		{
			"with nested enum",
			&MessageWithNestedEnum{
				Field: Nested_DEF,
			},
			"{\"field\":\"DEF\"}",
		},

		{
			"with deeply nested enum",
			&MessageWithDeeplyNestedEnum{
				Field: Deeply_Nested_DEF,
			},
			"{\"field\":\"DEF\"}",
		},
	}

	for _, tt := range jsonMarshalCases {
		t.Run(tt.Name, func(t *testing.T) {
			require := require.New(t)

			bs, err := json.Marshal(tt.Value)

			require.NoError(err)
			require.NotEmpty(bs)
			require.Equal(tt.Expected, string(bs))

			val := reflect.New(reflect.ValueOf(tt.Value).Elem().Type())

			require.NoError(json.Unmarshal(bs, val.Interface()))
			require.Equal(val.Interface(), tt.Value)
		})
	}

	var jsonUnmarshalCases = []struct {
		Name     string
		Value    string
		Expected interface{}
	}{
		{
			"match snake case with enum prefix",
			"{\"field\":\"DEF\"}",
			&ScreamingSnakeWithPrefxEnum{
				Field: ScreamingSnakeWithPrefix_SCREAMING_SNAKE_WITH_PREFIX_DEF,
			},
		},
	}

	for _, tt := range jsonUnmarshalCases {
		t.Run(tt.Name, func(t *testing.T) {
			require := require.New(t)
			val := reflect.New(reflect.ValueOf(tt.Expected).Elem().Type())

			require.NoError(json.Unmarshal([]byte(tt.Value), val.Interface()))
			require.Equal(tt.Expected, val.Interface())
		})
	}
}
