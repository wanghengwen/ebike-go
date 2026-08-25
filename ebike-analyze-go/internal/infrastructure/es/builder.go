package es

// M is a shorthand for map[string]interface{}, used to build JSON DSL bodies.
type M = map[string]interface{}

// Term builds a term query clause.
//
//	{"term": {field: value}}
func Term(field string, value interface{}) M {
	return M{"term": M{field: value}}
}

// Terms builds a terms query clause.
//
//	{"terms": {field: values}}
func Terms(field string, values interface{}) M {
	return M{"terms": M{field: values}}
}

// Prefix builds a prefix query clause.
//
//	{"prefix": {field: value}}
func Prefix(field string, value interface{}) M {
	return M{"prefix": M{field: value}}
}

// Exists builds an exists query clause.
//
//	{"exists": {"field": field}}
func Exists(field string) M {
	return M{"exists": M{"field": field}}
}

// RangeGte builds a range query with gte condition.
//
//	{"range": {field: {"gte": value}}}
func RangeGte(field string, value interface{}) M {
	return M{"range": M{field: M{"gte": value}}}
}

// RangeLte builds a range query with lte condition.
//
//	{"range": {field: {"lte": value}}}
func RangeLte(field string, value interface{}) M {
	return M{"range": M{field: M{"lte": value}}}
}

// RangeGteLte builds a range query with both gte and lte conditions.
//
//	{"range": {field: {"gte": gte, "lte": lte}}}
func RangeGteLte(field string, gte, lte interface{}) M {
	return M{"range": M{field: M{"gte": gte, "lte": lte}}}
}

// RangeGt builds a range query with gt condition.
//
//	{"range": {field: {"gt": value}}}
func RangeGt(field string, value interface{}) M {
	return M{"range": M{field: M{"gt": value}}}
}

// BoolMust builds a bool query with must clauses.
//
//	{"bool": {"must": clauses}}
func BoolMust(clauses ...M) M {
	return M{"bool": M{"must": clauses}}
}

// BoolShould builds a bool query with should clauses.
//
//	{"bool": {"should": clauses}}
func BoolShould(clauses ...M) M {
	return M{"bool": M{"should": clauses}}
}

// BoolMustNot builds a bool query with must_not clauses.
//
//	{"bool": {"must_not": clauses}}
func BoolMustNot(clauses ...M) M {
	return M{"bool": M{"must_not": clauses}}
}

// BoolQuery builds a bool query with optional must, should, and must_not clauses.
// Nil slices are omitted from the output.
func BoolQuery(must, should, mustNot []M) M {
	boolClause := M{}
	if len(must) > 0 {
		boolClause["must"] = must
	}
	if len(should) > 0 {
		boolClause["should"] = should
	}
	if len(mustNot) > 0 {
		boolClause["must_not"] = mustNot
	}
	return M{"bool": boolClause}
}

// ConstantScore builds a constant_score query wrapping the given filter.
//
//	{"constant_score": {"filter": filter}}
func ConstantScore(filter M) M {
	return M{"constant_score": M{"filter": filter}}
}
