// SPDX-License-Identifier: EUPL-1.2

package core

func TestError_ErrorLog_logger_Good(t *T) {
	log := NewLog(LogOptions{Level: LevelDebug, Output: NewBuffer()})

	AssertSame(t, log, (&ErrorLog{log: log}).logger())
}
func TestError_ErrorLog_logger_Bad(t *T) {
	AssertSame(t, Default(), (&ErrorLog{}).logger())
}
func TestError_ErrorLog_logger_Ugly(t *T) {
	original := Default()
	defer SetDefault(original)
	log := NewLog(LogOptions{Level: LevelQuiet, Output: NewBuffer()})
	SetDefault(log)

	AssertSame(t, log, (&ErrorLog{}).logger())
}
func TestError_ErrorPanic_appendReport_Good(t *T) {
	handler := &ErrorPanic{filePath: Path(t.TempDir(), "crash.json")}

	handler.appendReport(ax7CrashReport("panic: agent failed"))
	r := handler.Reports(0)

	RequireTrue(t, r.OK)
	reports := r.Value.([]CrashReport)
	AssertLen(t, reports, 1)
	AssertEqual(t, "panic: agent failed", reports[0].Error)
}
func TestError_ErrorPanic_appendReport_Bad(t *T) {
	path := Path(t.TempDir(), "crash.json")
	RequireTrue(t, WriteFile(path, []byte("{"), 0o600).OK)
	handler := &ErrorPanic{filePath: path}

	handler.appendReport(ax7CrashReport("panic: recovered"))
	r := handler.Reports(0)

	RequireTrue(t, r.OK)
	reports := r.Value.([]CrashReport)
	AssertLen(t, reports, 1)
	AssertEqual(t, "panic: recovered", reports[0].Error)
}
func TestError_ErrorPanic_appendReport_Ugly(t *T) {
	handler := &ErrorPanic{filePath: Path(t.TempDir(), "nested", "crash.json")}

	handler.appendReport(ax7CrashReport("panic: nested path"))
	r := handler.Reports(0)

	RequireTrue(t, r.OK)
	reports := r.Value.([]CrashReport)
	AssertLen(t, reports, 1)
	AssertEqual(t, "panic: nested path", reports[0].Error)
}
