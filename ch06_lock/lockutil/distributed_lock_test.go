package lockutil

import (
	"fmt"
	"github.com/google/uuid"
	"testing"
)

func TestName(t *testing.T) {
	uu := uuid.New().String()
	
	fmt.Println(uu)
	
	u,err:=uuid.NewUUID()
	fmt.Println(u.String())
	fmt.Println(err)
	
}
