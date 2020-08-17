// Copyright © 2011-12 Qtrac Ltd.
//
// This program or package and any associated files are licensed under the
// Apache License, Version 2.0 (the "License"); you may not use these files
// except in compliance with the License. You can get a copy of the License
// at: http://www.apache.org/licenses/LICENSE-2.0.
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package go_utils

type safeSlice chan scommandData

type scommandData struct {
	action  scommandAction
	index   int
	item    interface{}
	result  chan<- interface{}
	data    chan<- []interface{}
	updater SUpdateFunc
}

type scommandAction int

const (
	sinsert scommandAction = iota
	sremove
	sat
	supdate
	send
	slength
)

type SUpdateFunc func(interface{}) interface{}

type SafeSlice interface {
	Append(interface{})      // Append the given item to the slice
	At(int) interface{}      // Return the item at the given index position
	Close() []interface{}    // Close the channel and return the slice
	Delete(int)              // Delete the item at the given index position
	Len() int                // Return the number of items in the slice
	Update(int, SUpdateFunc) // Update the item at the given index position
}

func NewSafeSlice() SafeSlice {
	slice := make(safeSlice)
	go slice.run()
	return slice
}

func (slice safeSlice) run() {
	list := make([]interface{}, 0)
	for command := range slice {
		switch command.action {
		case sinsert:
			list = append(list, command.item)
		case sremove: // potentially expensive for long lists
			if 0 <= command.index && command.index < len(list) {
				list = append(list[:command.index],
					list[command.index+1:]...)
			}
		case sat:
			if 0 <= command.index && command.index < len(list) {
				command.result <- list[command.index]
			} else {
				command.result <- nil
			}
		case slength:
			command.result <- len(list)
		case supdate:
			if 0 <= command.index && command.index < len(list) {
				list[command.index] = command.updater(list[command.index])
			}
		case send:
			close(slice)
			command.data <- list
		}
	}
}

func (slice safeSlice) Append(item interface{}) {
	slice <- scommandData{action: sinsert, item: item}
}

func (slice safeSlice) Delete(index int) {
	slice <- scommandData{action: sremove, index: index}
}

func (slice safeSlice) At(index int) interface{} {
	reply := make(chan interface{})
	slice <- scommandData{sat, index, nil, reply, nil, nil}
	return <-reply
}

func (slice safeSlice) Len() int {
	reply := make(chan interface{})
	slice <- scommandData{action: slength, result: reply}
	return (<-reply).(int)
}

// If the updater calls a safeSlice method we will get deadlock!
func (slice safeSlice) Update(index int, updater SUpdateFunc) {
	slice <- scommandData{action: supdate, index: index, updater: updater}
}

func (slice safeSlice) Close() []interface{} {
	reply := make(chan []interface{})
	slice <- scommandData{action: send, data: reply}
	return <-reply
}
