package mongo

import (
	"context"
	"encoding/binary"
	"net"
	"time"

	"go.keploy.io/server/v2/pkg/core/proxy/integrations"
	"go.keploy.io/server/v2/utils"

	"go.keploy.io/server/v2/pkg/core/proxy/util"
	"go.keploy.io/server/v2/pkg/models"
	"go.mongodb.org/mongo-driver/x/mongo/driver/wiremessage"
	"go.uber.org/zap"
)

func init() {
	integrations.Register("mongo", NewMongo)
}

// TODO: Remove these global variables, and find a better way to handle this
var configRequests = []string{""}

type Mongo struct {
	logger *zap.Logger
}

func NewMongo(logger *zap.Logger) integrations.Integrations {
	return &Mongo{
		logger: logger,
	}
}

// MatchType determines if the outgoing network call is Mongo by comparing the
// message format with that of a mongo wire message.
func (m *Mongo) MatchType(_ context.Context, buffer []byte) bool {
	if len(buffer) < 4 {
		return false
	}
	messageLength := binary.LittleEndian.Uint32(buffer[0:4])
	return int(messageLength) == len(buffer)
}

func (m *Mongo) RecordOutgoing(ctx context.Context, src net.Conn, dst net.Conn, mocks chan<- *models.Mock, opts models.OutgoingOptions) error {
	logger := m.logger.With(zap.Any("Client IP Address", src.RemoteAddr().String()), zap.Any("Client ConnectionID", ctx.Value(models.ClientConnectionIDKey).(string)), zap.Any("Destination ConnectionID", ctx.Value(models.DestConnectionIDKey).(string)))
	reqBuf, err := util.ReadInitialBuf(ctx, logger, src)
	if err != nil {
		utils.LogError(logger, err, "failed to read the initial mongo message")
		return err
	}

	err = encodeMongo(ctx, logger, reqBuf, src, dst, mocks, opts)
	if err != nil {
		utils.LogError(logger, err, "failed to encode the mongo message into the yaml")
		return err
	}
	return nil
}

func (m *Mongo) MockOutgoing(ctx context.Context, src net.Conn, dstCfg *integrations.ConditionalDstCfg, mockDb integrations.MockMemDb, opts models.OutgoingOptions) error {
	logger := m.logger.With(zap.Any("Client IP Address", src.RemoteAddr().String()), zap.Any("Client ConnectionID", util.GetNextID()), zap.Any("Destination ConnectionID", util.GetNextID()))
	reqBuf, err := util.ReadInitialBuf(ctx, logger, src)
	if err != nil {
		utils.LogError(logger, err, "failed to read the initial mongo message")
		return err
	}

	err = decodeMongo(ctx, logger, reqBuf, src, dstCfg, mockDb, opts)
	if err != nil {
		utils.LogError(logger, err, "failed to decode the mongo message")
		return err
	}
	return nil
}

func recordMessage(_ context.Context, logger *zap.Logger, mongoRequests []models.MongoRequest, mongoResponses []models.MongoResponse, opReq Operation, reqTimestampMock time.Time, mocks chan<- *models.Mock) {
	// capture if the wiremessage is a mongo operation call

	shouldRecordCalls := true
	name := "mocks"
	meta1 := map[string]string{
		"operation": opReq.String(),
	}

	// Skip heartbeat from capturing in the global set of mocks. Since, the heartbeat packet always contain the "hello" boolean.
	// See: https://github.com/mongodb/mongo-go-driver/blob/8489898c64a2d8c2e2160006eb851a11a9db9e9d/x/mongo/driver/operation/hello.go#L503
	if isHeartBeat(logger, opReq, *mongoRequests[0].Header, mongoRequests[0].Message) {
		meta1["type"] = "config"
		for _, v := range configRequests {
			for _, req := range mongoRequests {

				switch req.Header.Opcode {
				case wiremessage.OpQuery:
					if req.Message.(*models.MongoOpQuery).Query == v {
						shouldRecordCalls = false
						break
					}
					configRequests = append(configRequests, req.Message.(*models.MongoOpQuery).Query)
				case wiremessage.OpMsg:
					if len(req.Message.(*models.MongoOpMessage).Sections) > 0 && req.Message.(*models.MongoOpMessage).Sections[0] == v {
						shouldRecordCalls = false
						break
					}
					configRequests = append(configRequests, req.Message.(*models.MongoOpMessage).Sections[0])
				default:
					if opReq.String() == v {
						shouldRecordCalls = false
						break
					}
					configRequests = append(configRequests, opReq.String())
				}
			}
		}
	}
	if shouldRecordCalls {
		mongoMock := &models.Mock{
			Version: models.GetVersion(),
			Kind:    models.Mongo,
			Name:    name,
			Spec: models.MockSpec{
				Metadata:         meta1,
				MongoRequests:    mongoRequests,
				MongoResponses:   mongoResponses,
				Created:          time.Now().Unix(),
				ReqTimestampMock: reqTimestampMock,
				ResTimestampMock: time.Now(),
			},
		}
		// Save the mock
		mocks <- mongoMock
	}
}
