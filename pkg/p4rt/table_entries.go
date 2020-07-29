package p4rt

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync/atomic"

	"github.com/golang/protobuf/proto"
	p4 "github.com/p4lang/p4runtime/go/p4/v1"
	"google.golang.org/grpc/codes"
)

var failedWrites uint32

func SendTableEntries(p4rt P4RuntimeClient, count uint64) {
	match := []*p4.FieldMatch{
		{
			FieldId:        1, // mpls_label
			FieldMatchType: &p4.FieldMatch_Exact_{&p4.FieldMatch_Exact{}},
		},
		// more fields...
		//{
		//	FieldId:        0,
		//	FieldMatchType: &p4.FieldMatch_Exact_{
		//		Exact: &p4.FieldMatch_Exact{[]byte{4, 5, 6, 7}}},
		//},
	}

	update := &p4.Update{
		Type: p4.Update_INSERT,
		Entity: &p4.Entity{Entity: &p4.Entity_TableEntry{
			TableEntry: &p4.TableEntry{
				TableId: 33574274, // FabricIngress.forwarding.mpls
				Match:   match,
				Action: &p4.TableAction{Type: &p4.TableAction_Action{Action: &p4.Action{
					ActionId: 16827758, // pop_mpls_and_next
					Params: []*p4.Action_Param{
						{
							ParamId: 1,              // next_id
							Value:   Uint64(0)[0:4], // 32 bits
						},
					},
				}}},
			},
		}},
	}

	for i := uint64(0); i < count; i++ {
		//update.GetEntity().GetTableEntry().GetMatch()[0].FieldId = uint32(i % 2)
		matchField := update.GetEntity().GetTableEntry().GetMatch()[0].GetExact()
		matchField.Value = Uint64(i)[5:8] // mpls_label is 20 bits
		res := p4rt.Write(update)
		go CountFailed(proto.Clone(update).(*p4.Update), res)
	}
}

func CountFailed(update *p4.Update, res <-chan *p4.Error) {
	err := <-res
	if err.CanonicalCode != int32(codes.OK) { // write failed
		atomic.AddUint32(&failedWrites, 1)
		fmt.Fprintf(os.Stderr, "%v -> %v\n", update, err.GetMessage())
	}
	//writeReples.Done()
}

func Uint64(v uint64) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, v)
	return bytes
}
