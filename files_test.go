package dskvs

import (
	"bytes"
	"os"
	"testing"
)

func TestWriteAndReadDirtyPage(t *testing.T) {
	expected := &page{
		isDirty:   true,
		isDeleted: false,
		key:       "Is that you, John Wayne? Is this me?",
		value: []byte(`Who said that? WHO THE FUCK said that?! Who's the slimy
commode of shit twinkle toed cocksucker just signed his own death warrant?!
Nobody huh? The fairy fucking godmother said it! Out fucking standing! I will
PT you all until you fucking DIE! I will PT you until your assholes are
sucking buttermilk!`),
	}

	filename := generateFilename(expected.key)
	writeToFile(expected)

	// Make sure to clean up if something goes wrong
	defer os.Remove(filename)
	if expected.isDirty {
		t.Errorf("Should have been cleaned up on write")
	}

	actual, err := readFromFile(filename)

	if err != nil {
		t.Fatalf("Failed reading file. %v", err)
	}

	if actual.key != expected.key {
		t.Errorf("Expected key <%s> but was <%s>",
			expected.key,
			actual.key)
	}

	expectedLen := len(expected.value)
	actualLen := len(actual.value)

	if actualLen != expectedLen {
		t.Errorf("Expected len(value) %d but was %d",
			expectedLen,
			actualLen)
	}

	var minLen int
	if expectedLen < actualLen {
		if expectedLen > 20 {
			minLen = 20
		} else {
			minLen = expectedLen
		}
	} else {
		if actualLen > 20 {
			minLen = 20
		} else {
			minLen = actualLen
		}
	}

	if !bytes.Equal(actual.value, expected.value) {
		t.Errorf("Expected value (truncated) <%v> but was <%v>",
			expected.value[:minLen],
			actual.value[:minLen])
	}
}

func TestDeletedPageDeleteFile(t *testing.T) {

	expected := &page{
		isDirty:   true,
		isDeleted: false,
		key:       "Is that you, John Wayne? Is this me?",
		value: []byte(`Who said that? WHO THE FUCK said that?! Who's the slimy
commode of shit twinkle toed cocksucker just signed his own death warrant?!
Nobody huh? The fairy fucking godmother said it! Out fucking standing! I will
PT you all until you fucking DIE! I will PT you until your assholes are
sucking buttermilk!`),
	}

	filename := generateFilename(expected.key)
	writeToFile(expected)
	defer os.Remove(filename)

	actual, err := readFromFile(filename)
	if err != nil {
		t.Errorf("Couldn't read page <%s> back : %v", filename, err)
	}

	actual.isDirty = true
	actual.isDeleted = true

	writeToFile(actual)

	if err = os.Remove(filename); os.IsExist(err) {
		t.Errorf("Didn't delete file <%s> : %v", filename, err)
	}

}
