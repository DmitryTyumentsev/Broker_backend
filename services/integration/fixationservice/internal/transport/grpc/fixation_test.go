package grpc

type mockService struct {
}

func newMockService() mockService {
	return mockService{}
}

//func TestNewFixation_ContextNotSend_CodesInvalidArgument(t *testing.T) {
//	h := NewHandler(fixationv1.UnimplementedFixationServiceServer{}, zap.NewNop())
//	req := &fixationv1.NewFixationRequest{
//		FixFor:    fixFor,
//		Phone:     randomPhone,
//		ProjectId: projectID,
//	}
//	resp, err := h.NewFixation(context.Background(), req)
//}
