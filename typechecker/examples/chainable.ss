// demonstrates chained member access calls resolved via receiver type lookup
class {
    data: #()
    start: false    // set to Date in Init; type becomes Date | False

    Init() {
        .start = Date.Begin()
    }

    // Object member chains
    Contains(x) {
        return .data.Has?(x)
    }

    IsMemberKnown(key) {
        return .data.Member?(key)
    }

    IsReadonly() {
        return .data.Readonly?()
    }

    // Date member chains: .start (Date | False) -> Date method -> typed result
    StartYear() {
        return .start.Year()
    }

    StartMonth() {
        return .start.Month()
    }

    StartFormatted(fmt) {
        return .start.ShortDate()
    }

    // deeper chain: .start.EndOfMonth() -> Date, then .ShortDate() -> String
    EndLabel() {
        return .start.EndOfMonth().ShortDate()
    }

    DaysFromStart(other) {
        return .start.MinusDays(other)
    }
}