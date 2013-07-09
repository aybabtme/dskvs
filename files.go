package dskvs

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
)

const (
	FILE_PERM = 0640
	DIR_PERM  = 0740
)

type fileHeader struct {
	Major         uint16
	Minor         uint16
	Patch         uint64
	Checksum      uint64
	KeyNameLength uint64
	PayloadLength uint64
}

var (
	fileHeaderSize int = binary.Size(new(fileHeader))
)

func newFileHeader(aPage *page) *fileHeader {
	hash := sha1.New().Sum(aPage.value)
	checksum, size := binary.Uvarint(hash)

	if size == 0 {
		log.Fatalf("Hash too small to produce uint64: %s", hash)
	} else if size < 0 {
		log.Printf("Hash to uint64 resulted in overflow. uint64=%d, overflow size=%d",
			checksum,
			size)
	}

	return &fileHeader{
		MAJOR_VERSION,
		MINOR_VERSION,
		PATCH_VERSION,
		checksum,
		uint64(len([]byte(aPage.key))),
		uint64(len(aPage.value)),
	}
}

func writeToFile(dirty *page) error {
	// Don't need to lock the page before reading the key, it's only modified
	// when `page` are created
	filename := generateFilename(dirty)
	// Lock the page for read
	dirty.RLock()
	if dirty.isDeleted {
		dirty.RUnlock()
		return deleteFile(filename)
	}

	data, err := fromPageToBytes(dirty)

	dirty.RUnlock()

	if err != nil {
		log.Printf("Couldn't get data from page: %v", err)
		return err
	}

	dirty.Lock()
	if dirty.isDeleted {
		// Was requested for deletion right after we tested
		dirty.Unlock()
		return writeToFile(dirty)
	}
	dirty.isDirty = false
	dirty.Unlock()

	if err := ioutil.WriteFile(filename, data, FILE_PERM); err != nil {
		log.Printf("Couldn't write file <%s> : %v", filename, err)
		return err
	}

	return nil
}

func readFromFile(filename string) (*page, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("Error reading file <%s> : %v", filename, err)
		return nil, err
	}

	header, err := headerFromBytes(data)
	if err != nil {
		log.Printf("Error reading header from file <%s> : %v",
			filename, err)
		return nil, errorCreatingHeader(filename, err)
	}

	// Fileformat is garanteed within same major versions
	if header.Major > MAJOR_VERSION {
		return nil, errorWrongVersion(header.Major, header.Minor, header.Patch)
	}

	keyIndex := uint64(fileHeaderSize)
	payloadIndex := keyIndex + header.KeyNameLength
	key := string(data[keyIndex:payloadIndex])
	payload := data[payloadIndex:]

	if uint64(len(data[payloadIndex:])) != header.PayloadLength {
		return nil, errorPayloadWrongSize(filename,
			header.PayloadLength,
			len(data[payloadIndex:]))
	}

	hash := sha1.New().Sum(payload)

	checksum, size := binary.Uvarint(hash)
	if size == 0 {
		log.Fatalf("Error reading file <%s> checksum, incomplete hash.",
			filename)
	} else if size < 0 {
		log.Fatal("Read too many bytes for checksum.")
	} else if checksum != header.Checksum {
		log.Printf("Payload checksum failed for file <%s>. Header says <%v>"+
			" but checksum was <%v>",
			filename,
			header,
			checksum)
		return nil, errorFailedChecksum(filename)
	}

	basepath := filepath.Base(filepath.Dir(filepath.Dir(filename)))
	coll := filepath.Base(filepath.Dir(filename))

	return &page{
		isDirty:   false,
		isDeleted: false,
		basepath:  basepath,
		coll:      coll,
		key:       key,
		value:     payload,
	}, nil
}

func deleteFile(filename string) error {
	err := os.Remove(filename)
	if os.IsNotExist(err) {
		// Duplicate request, or file doesn't exist.
		// Common case and not worth logging.
	} else if err != nil {
		log.Printf("Couldn't delete file <%s> : %v", filename, err)
		return err
	}
	return nil
}

func createFolder(create *member) error {
	folderName := filepath.Join(create.basepath, create.coll)
	if err := os.MkdirAll(folderName, DIR_PERM); err != nil {
		log.Printf("Couldn't create directory <%s> : %v", folderName, err)
		return err
	}
	return nil
}

func deleteFolder(delete *member) error {
	folderName := filepath.Join(delete.basepath, delete.coll)
	if err := os.RemoveAll(folderName); err != nil {
		log.Printf("Couldn't delete folder and children at <%s> : %v",
			folderName, err)
		return err
	}
	return nil
}

/*
	Helpers
*/

func headerFromBytes(data []byte) (*fileHeader, error) {
	var header fileHeader
	r := bytes.NewBuffer(data)
	err := binary.Read(r, binary.BigEndian, &header)
	if err != nil {
		return nil, err
	}
	return &header, err
}

func headerToBytes(header *fileHeader) ([]byte, error) {
	w := new(bytes.Buffer)
	err := binary.Write(w, binary.BigEndian, header)
	if err != nil {
		log.Printf("Error writing header to bytes : %v", err)
		return nil, err
	}
	return w.Bytes(), nil
}

func generateFilename(aPage *page) string {

	// url.QueryEscape incidentally escapes runes unsafe for a generateFilename
	escaped := url.QueryEscape(aPage.key)

	// Keep 40 first bytes for readability of generateFilename
	var max_length int
	if len(escaped) > 40 {
		max_length = 40
	} else {
		max_length = len(escaped)
	}

	prefix := []byte(escaped)[:max_length]

	// Append checksum value to the end, avoids collisions
	hash := sha1.New()
	_, _ = hash.Write([]byte(aPage.key))
	suffix := hex.EncodeToString(hash.Sum(nil))

	return filepath.Join(aPage.basepath, aPage.coll, string(prefix)+suffix)
}

func fromPageToBytes(aPage *page) ([]byte, error) {
	keyBytes := []byte(aPage.key)

	header := newFileHeader(aPage)
	headerBytes, err := headerToBytes(header)
	if err != nil {
		return nil, err
	}

	dataLength := len(headerBytes) + len(keyBytes) + len(aPage.value)
	keyIndex := len(headerBytes)
	payloadIndex := len(headerBytes) + len(keyBytes)

	data := make([]byte, dataLength)

	// Put the header
	copy(data, headerBytes)
	// Followed by the key name
	copy(data[keyIndex:], keyBytes)
	// Followed by the page value
	copy(data[payloadIndex:], aPage.value)

	return data, nil
}
