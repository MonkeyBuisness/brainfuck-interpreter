package bf

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_NewError(t *testing.T) {
	customErr := errors.New("error")

	err := NewError(ErrReadSymbol, customErr)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrReadSymbol))
	require.EqualError(t, err, fmt.Sprintf("%v: %v", ErrReadSymbol, customErr))
}
