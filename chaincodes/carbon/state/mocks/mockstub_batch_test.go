package mocks

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteBatchPutAndReadAfterFinish(t *testing.T) {
	stub := NewMockStub("carbon", nil)
	stub.MockTransactionStart("tx-batch-put")
	defer stub.MockTransactionEnd("tx-batch-put")

	stub.StartWriteBatch()

	err := stub.PutState("public-key", []byte("public-value"))
	require.NoError(t, err)
	err = stub.PutPrivateData("collectionA", "private-key", []byte("private-value"))
	require.NoError(t, err)

	// Buffered writes should not be visible before batch finish.
	publicBeforeFinish, err := stub.GetState("public-key")
	require.NoError(t, err)
	require.Nil(t, publicBeforeFinish)
	privateBeforeFinish, err := stub.GetPrivateData("collectionA", "private-key")
	require.NoError(t, err)
	require.Nil(t, privateBeforeFinish)

	err = stub.FinishWriteBatch()
	require.NoError(t, err)

	publicAfterFinish, err := stub.GetState("public-key")
	require.NoError(t, err)
	require.Equal(t, []byte("public-value"), publicAfterFinish)

	privateAfterFinish, err := stub.GetPrivateData("collectionA", "private-key")
	require.NoError(t, err)
	require.Equal(t, []byte("private-value"), privateAfterFinish)
}

func TestWriteBatchDeleteAndValidationParameter(t *testing.T) {
	stub := NewMockStub("carbon", nil)
	stub.MockTransactionStart("tx-batch-delete")
	defer stub.MockTransactionEnd("tx-batch-delete")

	err := stub.PutState("delete-key", []byte("to-delete"))
	require.NoError(t, err)
	err = stub.PutPrivateData("collectionB", "delete-private-key", []byte("to-delete-private"))
	require.NoError(t, err)

	stub.StartWriteBatch()
	err = stub.DelState("delete-key")
	require.NoError(t, err)
	err = stub.DelPrivateData("collectionB", "delete-private-key")
	require.NoError(t, err)
	err = stub.SetStateValidationParameter("meta-key", []byte("ep-public"))
	require.NoError(t, err)
	err = stub.SetPrivateDataValidationParameter("collectionB", "meta-private-key", []byte("ep-private"))
	require.NoError(t, err)

	err = stub.FinishWriteBatch()
	require.NoError(t, err)

	publicAfterDelete, err := stub.GetState("delete-key")
	require.NoError(t, err)
	require.Nil(t, publicAfterDelete)
	privateAfterDelete, err := stub.GetPrivateData("collectionB", "delete-private-key")
	require.NoError(t, err)
	require.Nil(t, privateAfterDelete)

	publicEP, err := stub.GetStateValidationParameter("meta-key")
	require.NoError(t, err)
	require.Equal(t, []byte("ep-public"), publicEP)

	privateEP, err := stub.GetPrivateDataValidationParameter("collectionB", "meta-private-key")
	require.NoError(t, err)
	require.Equal(t, []byte("ep-private"), privateEP)
}
