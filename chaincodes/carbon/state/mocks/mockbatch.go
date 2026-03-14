package mocks

type mockBatchRecordType int

const (
	mockBatchDataKeyType mockBatchRecordType = iota
	mockBatchMetadataKeyType
)

type mockBatchOperationType int

const (
	mockBatchPutState mockBatchOperationType = iota
	mockBatchDelState
	mockBatchPutValidationParameter
)

type mockBatchKey struct {
	Collection string
	Key        string
	Type       mockBatchRecordType
}

type mockWriteRecord struct {
	Operation mockBatchOperationType
	Value     []byte
}

type mockWriteBatch struct {
	writes map[mockBatchKey]mockWriteRecord
}

func newMockWriteBatch() *mockWriteBatch {
	return &mockWriteBatch{
		writes: make(map[mockBatchKey]mockWriteRecord),
	}
}

func (b *mockWriteBatch) putState(collection, key string, value []byte) {
	if b == nil {
		return
	}
	b.writes[mockBatchKey{
		Collection: collection,
		Key:        key,
		Type:       mockBatchDataKeyType,
	}] = mockWriteRecord{
		Operation: mockBatchPutState,
		Value:     value,
	}
}

func (b *mockWriteBatch) delState(collection, key string) {
	if b == nil {
		return
	}
	b.writes[mockBatchKey{
		Collection: collection,
		Key:        key,
		Type:       mockBatchDataKeyType,
	}] = mockWriteRecord{
		Operation: mockBatchDelState,
	}
}

func (b *mockWriteBatch) putValidationParameter(collection, key string, ep []byte) {
	if b == nil {
		return
	}
	b.writes[mockBatchKey{
		Collection: collection,
		Key:        key,
		Type:       mockBatchMetadataKeyType,
	}] = mockWriteRecord{
		Operation: mockBatchPutValidationParameter,
		Value:     ep,
	}
}
