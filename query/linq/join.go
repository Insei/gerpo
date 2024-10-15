package linq

//func NewJoinBuilder(core *CoreBuilder) *JoinBuilder {
//	return &JoinBuilder{
//		core: core,
//	}
//}
//
//type JoinBuilder struct {
//	core *CoreBuilder
//	opts []func(*sql.StringGroupBuilder)
//}
//
//func (q *JoinBuilder) Apply(b *sql.StringGroupBuilder) {
//	for _, opt := range q.opts {
//		opt(b)
//	}
//}
//
//func (q *JoinBuilder) LeftJoin(leftJoinFn func(ctx context.Context) string) {
//	for _, fieldPtr := range fieldsPtr {
//		col := q.core.GetColumn(fieldPtr)
//		q.opts = append(q.opts, func(b *sql.StringGroupBuilder) {
//			b.GroupBy(col)
//		})
//	}
//}
