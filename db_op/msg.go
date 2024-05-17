package db_op

type SqlQuery struct {
	Sql  string
	Args []interface{}
}

type SqlMultiQuery struct {
	SqlQuery []*SqlQuery
}
