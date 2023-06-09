package configurable

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfigurable(t *testing.T) {
	os.Clearenv()

	conf := NewConfigurable()

	t.Run("test NewInt and Int", func(t *testing.T) {
		// Register an integer flag with a default value
		conf.NewInt("test_int", 100, "integer test")
		// Get the pointer to the flag value
		p := conf.Int("test_int")
		// Check that the pointer is not nil and the default value is correct
		assert.NotNil(t, p)
		assert.Equal(t, 100, *p)
	})

	t.Run("test NewString and String", func(t *testing.T) {
		conf.NewString("test_string", "default", "string test")
		p := conf.String("test_string")
		assert.NotNil(t, p)
		assert.Equal(t, "default", *p)
	})

	t.Run("test NewBool and Bool", func(t *testing.T) {
		conf.NewBool("test_bool", false, "bool test")
		p := conf.Bool("test_bool")
		assert.NotNil(t, p)
		assert.Equal(t, false, *p)
	})

	t.Run("test NewFloat64 and Float64", func(t *testing.T) {
		conf.NewFloat64("test_float64", 123.456, "float64 test")
		p := conf.Float64("test_float64")
		assert.NotNil(t, p)
		assert.Equal(t, 123.456, *p)
	})

	t.Run("test NewDuration and Duration", func(t *testing.T) {
		conf.NewDuration("test_duration", 5*time.Minute, "duration test")
		p := conf.Duration("test_duration")
		assert.NotNil(t, p)
		assert.Equal(t, 5*time.Minute, *p)
	})

	t.Run("test NewInt64 and Int64", func(t *testing.T) {
		conf.NewInt64("test_int64", 64, "int64 test")
		p := conf.Int64("test_int64")
		assert.NotNil(t, p)
		assert.Equal(t, int64(64), *p)
	})

	t.Run("test Parse", func(t *testing.T) {
		err := conf.Parse("")
		assert.NoError(t, err)
	})

	t.Run("test LoadFile", func(t *testing.T) {
		// Create a temporary config file, write some config to it, then use that for the test
		err := conf.LoadFile("unknown_file_type.unknown")
		assert.Error(t, err)
	})

	t.Run("test Usage", func(t *testing.T) {
		usage := conf.Usage()
		assert.Contains(t, usage, "test_int")
		assert.Contains(t, usage, "test_string")
		assert.Contains(t, usage, "test_bool")
		assert.Contains(t, usage, "test_float64")
		assert.Contains(t, usage, "test_duration")
		assert.Contains(t, usage, "test_int64")
	})
}
