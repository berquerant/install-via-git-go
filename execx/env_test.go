package execx_test

import (
	"berquerant/install-via-git-go/execx"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnv(t *testing.T) {

	t.Run("set", func(t *testing.T) {
		e := execx.NewEnv()
		{
			_, ok := e.Get("KEY")
			assert.False(t, ok)
		}
		e.Set("KEY", "VAL")
		{
			got, ok := e.Get("KEY")
			assert.True(t, ok)
			assert.Equal(t, "VAL", got)
		}
		assert.Equal(t, []string{"KEY=VAL"}, e.IntoSlice())
	})

	t.Run("expand", func(t *testing.T) {
		e := execx.EnvFromSlice([]string{"KEY=VAL"})
		assert.Equal(t, "value is VAL", e.Expand("value is $KEY"))

		e.Set("KEY2", "new${KEY}")
		assert.Equal(t, "value is newVAL", e.Expand("value is $KEY2"))
	})

	t.Run("expand exhausted", func(t *testing.T) {
		e := execx.EnvFromSlice([]string{"A=B", "B=C", "C=A"})
		t.Log(e.Expand("value is $A"))
	})

	t.Run("expand strings", func(t *testing.T) {
		e := execx.EnvFromSlice([]string{"SUBJECT=Alice", "LOCATION1=Virginia", "LOCATION2=Atlanta"})
		assert.Equal(t, []string{
			"Alice went to Virginia",
			"Alice went to Atlanta",
		}, e.ExpandStrings([]string{
			"$SUBJECT went to $LOCATION1",
			"$SUBJECT went to $LOCATION2",
		}))
	})
}
