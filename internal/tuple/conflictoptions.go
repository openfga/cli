package tuple

import (
	"fmt"

	"github.com/openfga/go-sdk/client"

	"github.com/openfga/cli/internal/clierrors"
)

type ClientWriteRequestOnDuplicateWrites client.ClientWriteRequestOnDuplicateWrites

const (
	CLIENT_WRITE_REQUEST_ON_DUPLICATE_WRITES_ERROR  ClientWriteRequestOnDuplicateWrites = "error"  //nolint:revive
	CLIENT_WRITE_REQUEST_ON_DUPLICATE_WRITES_IGNORE ClientWriteRequestOnDuplicateWrites = "ignore" //nolint:revive
)

func (option *ClientWriteRequestOnDuplicateWrites) String() string {
	return string(*option)
}

func (option *ClientWriteRequestOnDuplicateWrites) ToSdkEnum() client.ClientWriteRequestOnDuplicateWrites {
	return client.ClientWriteRequestOnDuplicateWrites(*option)
}

func (option *ClientWriteRequestOnDuplicateWrites) Set(v string) error {
	switch v {
	case "error", "ignore":
		*option = ClientWriteRequestOnDuplicateWrites(v)

		return nil
	default:
		return fmt.Errorf(
			`%w: must be one of "%v" or "%v"`,
			clierrors.ErrInvalidFormat,
			CLIENT_WRITE_REQUEST_ON_DUPLICATE_WRITES_ERROR,
			CLIENT_WRITE_REQUEST_ON_DUPLICATE_WRITES_IGNORE,
		)
	}
}

func (option *ClientWriteRequestOnDuplicateWrites) Type() string {
	return "error|ignore"
}

type ClientWriteRequestOnMissingDeletes client.ClientWriteRequestOnMissingDeletes

const (
	CLIENT_WRITE_REQUEST_ON_MISSING_DELETES_ERROR  ClientWriteRequestOnMissingDeletes = "error"  //nolint:revive
	CLIENT_WRITE_REQUEST_ON_MISSING_DELETES_IGNORE ClientWriteRequestOnMissingDeletes = "ignore" //nolint:revive
)

func (option *ClientWriteRequestOnMissingDeletes) String() string {
	return string(*option)
}

func (option *ClientWriteRequestOnMissingDeletes) ToSdkEnum() client.ClientWriteRequestOnMissingDeletes {
	return client.ClientWriteRequestOnMissingDeletes(*option)
}

func (option *ClientWriteRequestOnMissingDeletes) Set(v string) error {
	switch v {
	case "error", "ignore":
		*option = ClientWriteRequestOnMissingDeletes(v)

		return nil
	default:
		return fmt.Errorf(
			`%w: must be one of "%v" or "%v"`,
			clierrors.ErrInvalidFormat,
			CLIENT_WRITE_REQUEST_ON_MISSING_DELETES_ERROR,
			CLIENT_WRITE_REQUEST_ON_MISSING_DELETES_IGNORE,
		)
	}
}

func (option *ClientWriteRequestOnMissingDeletes) Type() string {
	return "error|ignore"
}
