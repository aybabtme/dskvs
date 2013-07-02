package dskvs

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var genericPage = &page{
	isDirty:   true,
	isDeleted: false,
	basepath:  "imdb",
	coll:      "FMJ",
	key:       "Is that you, John Wayne? Is this me?",
	value: []byte(`Who said that? WHO THE FUCK said that?! Who's the slimy
commode of shit twinkle toed cocksucker just signed his own death warrant?!
Nobody huh? The fairy fucking godmother said it! Out fucking standing! I will
PT you all until you fucking DIE! I will PT you until your assholes are
sucking buttermilk!`),
}

func TestWriteAndReadDirtyPage(t *testing.T) {
	expected := genericPage

	filename := generateFilename(expected)
	os.MkdirAll(filepath.Dir(filename), DIR_PERM)
	defer os.RemoveAll(expected.basepath)

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

func TestDeletingPageShouldDeleteFile(t *testing.T) {

	expected := genericPage

	filename := generateFilename(expected)
	os.MkdirAll(filepath.Dir(filename), DIR_PERM)
	defer os.RemoveAll(expected.basepath)

	writeToFile(expected)

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

func TestErrorWhenReadingJunkFile(t *testing.T) {
	filename := "junk_file.test"
	ioutil.WriteFile(filename, []byte{0xDE, 0xAD, 0xBE, 0xEF}, FILE_PERM)
	defer os.Remove(filename)

	_, err := readFromFile(filename)
	if _, isRightType := err.(FileError); !isRightType {
		t.Errorf("Should have returned an error of type FileError"+
			", error was %v",
			err)
	}
	err.Error() // Call it to make gocov happy
}

func TestErrorWhenOpeningDifferentMajorVersion(t *testing.T) {
	filename := "incompatible_version.test"
	aPage := genericPage

	// Get the header as it will be written to disk
	currentHeader := newFileHeader(aPage)
	// Modify it
	currentHeader.Major = MAJOR_VERSION + 1
	headerBytes, err := headerToBytes(currentHeader)
	if err != nil {
		t.Fatalf("Couldn't get fake header, %v", err)
	}

	// Get the data
	pageBytes, err := fromPageToBytes(aPage)
	// Overwrite the header
	copy(pageBytes, headerBytes)

	if err := ioutil.WriteFile(filename, pageBytes, FILE_PERM); err != nil {
		t.Fatalf("Couldn't write file <%s> : %v", filename, err)
	}
	defer os.Remove(filename)

	_, err = readFromFile(filename)
	if _, isRightType := err.(FileError); !isRightType {
		t.Errorf("Should have returned an error of type FileError"+
			", error was %v",
			err)
	}
	err.Error() // Call it to make gocov happy

}

func TestErrorWhenHeaderComesFromAnotherPageWithSimilarLength(t *testing.T) {
	filename := "impostor.test"
	aPage := genericPage

	impostor := &page{
		isDirty:   true,
		isDeleted: false,
		basepath:  "imdb",
		coll:      "FMJ",
		key:       "Is that you, John Wayne? Is this me?",
		// value has same length but different text
		value: []byte(`HAHAHA yes !!? WHO THE FUCK said that?! Who's the slimy
commode of shit twinkle toed cocksucker just signed his own death warrant?!
Nobody huh? The fairy fucking godmother said it! Out fucking standing! I will
PT you all until you fucking DIE! I will PT you until your assholes are
sucking buttermilk!`),
	}

	impostorHeader := newFileHeader(impostor)
	headerBytes, err := headerToBytes(impostorHeader)
	if err != nil {
		t.Fatalf("Couldn't get fake header, %v", err)
	}
	pageBytes, err := fromPageToBytes(aPage)
	copy(pageBytes, headerBytes)

	if err := ioutil.WriteFile(filename, pageBytes, FILE_PERM); err != nil {
		t.Fatalf("Couldn't write file <%s> : %v", filename, err)
	}
	defer os.Remove(filename)

	_, err = readFromFile(filename)
	if _, isRightType := err.(FileError); !isRightType {
		t.Errorf("Should have returned an error of type FileError"+
			", error was %v",
			err)
	}
	err.Error() // Call it to make gocov happy
}

func TestErrorWhenHeaderComesFromAnotherPageWithDifferentLength(t *testing.T) {
	filename := "impostor.test"
	aPage := genericPage

	pageBytes, err := fromPageToBytes(aPage)

	impostor := genericPage
	impostor.value = []byte("hahahaha yes it's me")

	impostorHeader := newFileHeader(impostor)
	headerBytes, err := headerToBytes(impostorHeader)
	if err != nil {
		t.Fatalf("Couldn't get fake header, %v", err)
	}

	copy(pageBytes, headerBytes)

	if err := ioutil.WriteFile(filename, pageBytes, FILE_PERM); err != nil {
		t.Fatalf("Couldn't write file <%s> : %v", filename, err)
	}
	defer os.Remove(filename)

	result, err := readFromFile(filename)
	if _, isRightType := err.(FileError); !isRightType {
		t.Errorf("Should have returned an error of type FileError"+
			", error was %v, page was : %v",
			err,
			result)
	}
	err.Error() // Call it to make gocov happy
}

func TestErrorWhenChecksumIsWrong(t *testing.T) {
	filename := "wrong_checksum.test"
	aPage := &page{
		isDirty:   true,
		isDeleted: false,
		basepath:  "imdb",
		coll:      "FMJ",
		key:       "Is that you, John Wayne? Is this me?",
		value: []byte(`Who said that? WHO THE FUCK said that?! Who's the slimy
commode of shit twinkle toed cocksucker just signed his own death warrant?!
Nobody huh? The fairy fucking godmother said it! Out fucking standing! I will
PT you all until you fucking DIE! I will PT you until your assholes are
sucking buttermilk!`),
	}

	currentHeader := newFileHeader(aPage)
	currentHeader.Checksum = currentHeader.Checksum + 1
	headerBytes, err := headerToBytes(currentHeader)
	if err != nil {
		t.Fatalf("Couldn't get fake header, %v", err)
	}
	pageBytes, err := fromPageToBytes(aPage)
	copy(pageBytes, headerBytes)

	if err := ioutil.WriteFile(filename, pageBytes, FILE_PERM); err != nil {
		t.Fatalf("Couldn't write file <%s> : %v", filename, err)
	}
	defer os.Remove(filename)

	_, err = readFromFile(filename)
	if _, isRightType := err.(FileError); !isRightType {
		t.Errorf("Should have returned an error of type FileError"+
			", error was %v",
			err)
	}
	err.Error() // Call it to make gocov happy
}
