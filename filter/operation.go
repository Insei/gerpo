package filter

type Operation string

const (
	// OperationEQ is a constant of type Operation that represents the operation where the field is equal to the value
	OperationEQ = Operation("eq")

	// OperationNEQ is a constant of type Operation that represents the operation where the field is not equal to the value.
	OperationNEQ = Operation("neq")

	// OperationGT is a constant of type Operation that represents the operation where the field is greater than the value.
	OperationGT = Operation("gt")

	// OperationGTE is a constant of type Operation and represents the operation where the field is greater than or equal to the value.
	OperationGTE = Operation("gte")

	// OperationLT is a constant of type Operation and represents the operation where the field is less than the value.
	OperationLT = Operation("lt")

	// OperationLTE is a constant of type Operation that represents the operation where the field is less than or equal to the value.
	OperationLTE = Operation("lte")

	// OperationIN is a constant of type Operation that represents the operation where the field is in the specified values.
	OperationIN = Operation("in")

	// OperationNIN is a constant of type Operation that represents the operation where the field is not in the specified values.
	OperationNIN = Operation("nin")

	// OperationEW is a constant of type Operation that represents the operation where the field ends with the value.
	OperationEW = Operation("ew")

	// OperationNEW is a constant of type Operation. It represents the operation where the field not ends with the value.
	OperationNEW = Operation("new")

	// OperationBW is a constant of type Operation. It represents the operation where the field begins with the value.
	OperationBW = Operation("bw")

	// OperationNBW is a constant of type Operation that represents the operation where the field begins with the value.
	OperationNBW = Operation("nbw")
)
