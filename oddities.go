/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mirbft

import (
	"fmt"

	pb "github.com/IBM/mirbft/mirbftpb"
	"go.uber.org/zap"
)

const (
	SeqNoLog   = "SeqNo"
	ReqNoLog   = "ReqNo"
	EpochLog   = "Epoch"
	NodeIDLog  = "NodeID"
	MsgTypeLog = "MsgType"
)

func logBasics(source NodeID, msg *pb.Msg) []zap.Field {
	fields := []zap.Field{
		zap.Uint64(NodeIDLog, uint64(source)),
	}

	switch innerMsg := msg.Type.(type) {
	case *pb.Msg_EpochChange:
		msg := innerMsg.EpochChange
		fields = append(fields,
			zap.String(MsgTypeLog, "epochchange"),
			zap.Uint64(EpochLog, msg.NewEpoch),
		)
	case *pb.Msg_Preprepare:
		msg := innerMsg.Preprepare
		fields = append(fields,
			zap.String(MsgTypeLog, "preprepare"),
			zap.Uint64(SeqNoLog, msg.SeqNo),
			zap.Uint64(EpochLog, msg.Epoch),
		)
	case *pb.Msg_Prepare:
		msg := innerMsg.Prepare
		fields = append(fields,
			zap.String(MsgTypeLog, "prepare"),
			zap.Uint64(SeqNoLog, msg.SeqNo),
			zap.Uint64(EpochLog, msg.Epoch),
		)
	case *pb.Msg_Commit:
		msg := innerMsg.Commit
		fields = append(fields,
			zap.String(MsgTypeLog, "commit"),
			zap.Uint64(SeqNoLog, msg.SeqNo),
			zap.Uint64(EpochLog, msg.Epoch),
		)
	case *pb.Msg_Checkpoint:
		msg := innerMsg.Checkpoint
		fields = append(fields,
			zap.String(MsgTypeLog, "checkpoint"),
			zap.Uint64(SeqNoLog, msg.SeqNo),
		)
	case *pb.Msg_ForwardRequest:
		msg := innerMsg.ForwardRequest
		fields = append(fields,
			zap.String(MsgTypeLog, "forwardrequest"),
			zap.Uint64(ReqNoLog, msg.Request.ReqNo),
			zap.Binary(ReqNoLog, msg.Request.ClientId),
		)
	default:
		fields = append(fields,
			zap.String(MsgTypeLog, fmt.Sprintf("%T", msg.Type)),
		)
	}

	return fields
}

// oddities are events which are not necessarily damaging
// or detrimental to the state machine, but which may represent
// byzantine behavior, misconfiguration, or bugs.
type oddities struct {
	logger Logger
	nodes  map[NodeID]*oddity
}

type oddity struct {
	invalid          uint64
	alreadyProcessed uint64
	// aboveWatermarks uint64
	// belowWatermarks uint64
	// wrongEpoch      uint64
}

func (o *oddities) getNode(nodeID NodeID) *oddity {
	if o.nodes == nil {
		o.nodes = map[NodeID]*oddity{}
	}

	od, ok := o.nodes[nodeID]
	if !ok {
		od = &oddity{}
		o.nodes[nodeID] = od
	}
	return od
}

func (o *oddities) alreadyProcessed(source NodeID, msg *pb.Msg) {
	o.logger.Debug("already processed message", logBasics(source, msg)...)
	o.getNode(source).alreadyProcessed++
}

/* // TODO enable again when we add back these checks
func (o *oddities) aboveWatermarks(source NodeID, msg *pb.Msg) {
	o.logger.Warn("received message above watermarks", logBasics(source, msg)...)
	o.getNode(source).aboveWatermarks++
}

func (o *oddities) belowWatermarks(source NodeID, msg *pb.Msg) {
	o.logger.Warn("received message below watermarks", logBasics(source, msg)...)
	o.getNode(source).belowWatermarks++
}

*/

func (o *oddities) invalidMessage(source NodeID, msg *pb.Msg) {
	o.logger.Error("invalid message", logBasics(source, msg)...)
	o.getNode(source).invalid++
}
