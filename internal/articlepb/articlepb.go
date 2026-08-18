package articlepb

import (
	"context"
	"encoding/json"
	"io"

	"google.golang.org/grpc"
)

// JSONCodec keeps internal service-to-service gRPC transport lightweight.
type JSONCodec struct{}

func (JSONCodec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONCodec) Unmarshal(data []byte, out any) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}

func (JSONCodec) Name() string {
	return "json"
}

type Post struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Category    string `json:"category"`
	Status      string `json:"status"`
	CreatedDate string `json:"created_date"`
	UpdatedDate string `json:"updated_date"`
}

type CreateArticleRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
	Status   string `json:"status"`
}

type CreateArticleResponse struct {
	ID      uint   `json:"id"`
	Message string `json:"message"`
}

type ListArticlesRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type ListArticlesResponse struct {
	Posts []*Post `json:"posts"`
}

type GetArticleRequest struct {
	Id uint `json:"id"`
}

type GetArticleResponse struct {
	Post *Post `json:"post"`
}

type UpdateArticleRequest struct {
	Id       uint   `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
	Status   string `json:"status"`
}

type UpdateArticleResponse struct {
	ID      uint   `json:"id"`
	Message string `json:"message"`
}

type DeleteArticleRequest struct {
	Id uint `json:"id"`
}

type DeleteArticleResponse struct {
	ID      uint   `json:"id"`
	Message string `json:"message"`
}

type SubscribeEventsRequest struct{}

type ArticleEvent struct {
	Type    string `json:"type"`
	Post    *Post  `json:"post"`
	EventAt string `json:"event_at"`
}

type ArticleServiceServer interface {
	CreateArticle(context.Context, *CreateArticleRequest) (*CreateArticleResponse, error)
	ListArticles(context.Context, *ListArticlesRequest) (*ListArticlesResponse, error)
	GetArticle(context.Context, *GetArticleRequest) (*GetArticleResponse, error)
	UpdateArticle(context.Context, *UpdateArticleRequest) (*UpdateArticleResponse, error)
	DeleteArticle(context.Context, *DeleteArticleRequest) (*DeleteArticleResponse, error)
	SubscribeEvents(*SubscribeEventsRequest, ArticleService_SubscribeEventsServer) error
}

type ArticleServiceClient interface {
	CreateArticle(context.Context, *CreateArticleRequest, ...grpc.CallOption) (*CreateArticleResponse, error)
	ListArticles(context.Context, *ListArticlesRequest, ...grpc.CallOption) (*ListArticlesResponse, error)
	GetArticle(context.Context, *GetArticleRequest, ...grpc.CallOption) (*GetArticleResponse, error)
	UpdateArticle(context.Context, *UpdateArticleRequest, ...grpc.CallOption) (*UpdateArticleResponse, error)
	DeleteArticle(context.Context, *DeleteArticleRequest, ...grpc.CallOption) (*DeleteArticleResponse, error)
	SubscribeEvents(context.Context, *SubscribeEventsRequest, ...grpc.CallOption) (ArticleService_SubscribeEventsClient, error)
}

type ArticleService_SubscribeEventsClient interface {
	Recv() (*ArticleEvent, error)
	grpc.ClientStream
}

type ArticleService_SubscribeEventsServer interface {
	Send(*ArticleEvent) error
	grpc.ServerStream
}

type articleServiceClient struct {
	cc grpc.ClientConnInterface
}

// NewArticleServiceClient builds a gRPC client for article-service.
func NewArticleServiceClient(cc grpc.ClientConnInterface) ArticleServiceClient {
	return &articleServiceClient{cc: cc}
}

func (c *articleServiceClient) CreateArticle(ctx context.Context, in *CreateArticleRequest, opts ...grpc.CallOption) (*CreateArticleResponse, error) {
	out := new(CreateArticleResponse)
	err := c.cc.Invoke(ctx, "/sharing.vision.article.v1.ArticleService/CreateArticle", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *articleServiceClient) ListArticles(ctx context.Context, in *ListArticlesRequest, opts ...grpc.CallOption) (*ListArticlesResponse, error) {
	out := new(ListArticlesResponse)
	err := c.cc.Invoke(ctx, "/sharing.vision.article.v1.ArticleService/ListArticles", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *articleServiceClient) GetArticle(ctx context.Context, in *GetArticleRequest, opts ...grpc.CallOption) (*GetArticleResponse, error) {
	out := new(GetArticleResponse)
	err := c.cc.Invoke(ctx, "/sharing.vision.article.v1.ArticleService/GetArticle", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *articleServiceClient) UpdateArticle(ctx context.Context, in *UpdateArticleRequest, opts ...grpc.CallOption) (*UpdateArticleResponse, error) {
	out := new(UpdateArticleResponse)
	err := c.cc.Invoke(ctx, "/sharing.vision.article.v1.ArticleService/UpdateArticle", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *articleServiceClient) DeleteArticle(ctx context.Context, in *DeleteArticleRequest, opts ...grpc.CallOption) (*DeleteArticleResponse, error) {
	out := new(DeleteArticleResponse)
	err := c.cc.Invoke(ctx, "/sharing.vision.article.v1.ArticleService/DeleteArticle", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *articleServiceClient) SubscribeEvents(ctx context.Context, in *SubscribeEventsRequest, opts ...grpc.CallOption) (ArticleService_SubscribeEventsClient, error) {
	stream, err := c.cc.NewStream(ctx, &_ArticleService_serviceDesc.Streams[0], "/sharing.vision.article.v1.ArticleService/SubscribeEvents", opts...)
	if err != nil {
		return nil, err
	}
	if err := stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := stream.CloseSend(); err != nil {
		return nil, err
	}
	return &articleServiceSubscribeEventsClient{stream}, nil
}

type articleServiceSubscribeEventsClient struct {
	grpc.ClientStream
}

func (c *articleServiceSubscribeEventsClient) Recv() (*ArticleEvent, error) {
	m := new(ArticleEvent)
	if err := c.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// RegisterArticleServiceServer registers the grpc handlers for article service.
func RegisterArticleServiceServer(s *grpc.Server, srv ArticleServiceServer) {
	s.RegisterService(&_ArticleService_serviceDesc, srv)
}

func _ArticleService_CreateArticle_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateArticleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ArticleServiceServer).CreateArticle(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sharing.vision.article.v1.ArticleService/CreateArticle",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ArticleServiceServer).CreateArticle(ctx, req.(*CreateArticleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ArticleService_ListArticles_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListArticlesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ArticleServiceServer).ListArticles(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sharing.vision.article.v1.ArticleService/ListArticles",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ArticleServiceServer).ListArticles(ctx, req.(*ListArticlesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ArticleService_GetArticle_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetArticleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ArticleServiceServer).GetArticle(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sharing.vision.article.v1.ArticleService/GetArticle",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ArticleServiceServer).GetArticle(ctx, req.(*GetArticleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ArticleService_UpdateArticle_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateArticleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ArticleServiceServer).UpdateArticle(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sharing.vision.article.v1.ArticleService/UpdateArticle",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ArticleServiceServer).UpdateArticle(ctx, req.(*UpdateArticleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ArticleService_DeleteArticle_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteArticleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ArticleServiceServer).DeleteArticle(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sharing.vision.article.v1.ArticleService/DeleteArticle",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ArticleServiceServer).DeleteArticle(ctx, req.(*DeleteArticleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ArticleService_SubscribeEvents_Handler(srv interface{}, stream grpc.ServerStream) error {
	in := new(SubscribeEventsRequest)
	if err := stream.RecvMsg(in); err != nil {
		if err != io.EOF {
			return err
		}
	}
	return srv.(ArticleServiceServer).SubscribeEvents(in, &articleServiceSubscribeEventsServer{ServerStream: stream})
}

type articleServiceSubscribeEventsServer struct {
	grpc.ServerStream
}

func (x *articleServiceSubscribeEventsServer) Send(m *ArticleEvent) error {
	return x.ServerStream.SendMsg(m)
}

var _ArticleService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "sharing.vision.article.v1.ArticleService",
	HandlerType: (*ArticleServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "CreateArticle", Handler: _ArticleService_CreateArticle_Handler},
		{MethodName: "ListArticles", Handler: _ArticleService_ListArticles_Handler},
		{MethodName: "GetArticle", Handler: _ArticleService_GetArticle_Handler},
		{MethodName: "UpdateArticle", Handler: _ArticleService_UpdateArticle_Handler},
		{MethodName: "DeleteArticle", Handler: _ArticleService_DeleteArticle_Handler},
	},
	Streams: []grpc.StreamDesc{
		{StreamName: "SubscribeEvents", Handler: _ArticleService_SubscribeEvents_Handler, ServerStreams: true},
	},
	Metadata: "article_service.proto",
}
