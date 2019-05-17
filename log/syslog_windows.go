// +build !linux

package log
import (
	"errors"
)


func SetNetLogHandler(net,addr,tag string,fmtr Format)(Handler, error){
	return nil,errors.New("window not support")

}