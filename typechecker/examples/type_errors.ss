// ReportBuilder: a moderately complicated class that mixes correct code
// with four distinct type errors that ConstraintGen/Solve can catch.
//
// Correctly typed regions are marked [OK].
// Type errors are marked [ERROR] with the violation described.
class
	{
	// typed class members - these seed concrete types into every method
	title: "Monthly Report"
	year: 2024
	maxRows: 100

	// [OK] Header builds a header by converting the numeric year to a string first.
	Header()
		{
		return .title $ " - " $ String(.year)
		}

	// [ERROR 1] BadHeader cats .year (Number) directly into a string expression.
	//   .year :: Number, but $ (Cat) requires String on every operand.
	//   ConstraintGen emits: operand 3 of $ (Cat) - got Number, required String.
	BadHeader()
		{
		return .title $ " - " $ .year
		}

	// [OK] DaysUntilDate safely uses a Date - no arithmetic, just string formatting.
	DaysUntilDate()
		{
		d = Date.Begin()
		return d.ShortDate()
		}

	// [ERROR 2] DaysRemaining uses Date() (which returns False|Date) as a Number.
	//   today :: False | Date (from the Date() annotation).
	//   Subtraction requires Number on both sides; neither False nor Date is Number.
	//   ConstraintGen emits: right operand of Sub - got False | Date, required Number.
	DaysRemaining()
		{
		today = Date()
		remaining = .maxRows - today
		return remaining
		}

	// [OK] CurrentMonthEnd chains annotated Date calls - every receiver matches.
	//   Date.Begin() -> Date, EndOfMonth[this: Date] -> Date, ShortDate[this: Date] -> String.
	CurrentMonthEnd()
		{
		d = Date.Begin()
		return d.EndOfMonth().ShortDate()
		}

	// [ERROR 3] ExtractHour calls Hour() on a Number, not a Date.
	//   .year :: Number.  Hour[this: Date] requires a Date receiver.
	//   ConstraintGen emits: receiver of .Hour (requires Date) - got Number.
	ExtractHour()
		{
		n = .year
		return n.Hour()
		}

	// [ERROR 4] Paginate calls Iter() on a Number - Object method on wrong type.
	//   .maxRows :: Number.  Iter[this: Object] requires an Object receiver.
	//   ConstraintGen emits: receiver of .Iter (requires Object) - got Number.
	Paginate()
		{
		n = .maxRows
		return n.Iter()
		}
	}
