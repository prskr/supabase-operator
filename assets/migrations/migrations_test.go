package migrations_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/prskr/supabase-operator/assets/migrations"
)

func TestMigrationScripts(t *testing.T) {
	scriptOrErr := migrations.MigrationScripts()

	var numberOfScripts int

	for script, err := range scriptOrErr {
		numberOfScripts++
		assert.NoError(t, err)
		assert.NotZero(t, script.FileName)
		assert.NotZero(t, script.Content)
		assert.NotZero(t, script.Hash)
	}

	assert.NotEqual(t, 0, numberOfScripts, "Expected more than 0 scripts but got %d", numberOfScripts)
	t.Logf("Found %d scripts", numberOfScripts)
}

func TestInitScripts(t *testing.T) {
	scriptOrErr := migrations.InitScripts()

	var numberOfScripts int

	for script, err := range scriptOrErr {
		numberOfScripts++
		assert.NoError(t, err)
		assert.NotZero(t, script.FileName)
		assert.NotZero(t, script.Content)
		assert.NotZero(t, script.Hash)
	}

	assert.NotEqual(t, 0, numberOfScripts, "Expected more than 0 scripts but got %d", numberOfScripts)

	t.Logf("Found %d scripts", numberOfScripts)
}
