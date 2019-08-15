package s1ap

import (
	"fmt"

	"github.com/ishidawataru/sctp"
)

func ReadHeader(conn *sctp.SCTPConn) error {
	buf := make([]byte, 4)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Read err", err.Error())
		return err
	}
	fmt.Println("Readfull", n)
	return nil
}
