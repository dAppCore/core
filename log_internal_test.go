// SPDX-License-Identifier: EUPL-1.2

package core

func TestLog_Log_log_Good(t *T) {
	out := NewBuffer()
	log := NewLog(LogOptions{Level: LevelInfo, Output: out})

	log.log(LevelInfo, "[INF]", "agent ready", "site", "homelab")

	AssertContains(t, out.String(), "[INF]")
	AssertContains(t, out.String(), `site="homelab"`)
}
func TestLog_Log_log_Bad(t *T) {
	out := NewBuffer()
	log := NewLog(LogOptions{Level: LevelInfo, Output: out})

	log.log(LevelInfo, "[INF]", "dangling key", "session")

	AssertContains(t, out.String(), "session=<nil>")
}
func TestLog_Log_log_Ugly(t *T) {
	out := NewBuffer()
	log := NewLog(LogOptions{Level: LevelInfo, Output: out, RedactKeys: []string{"token"}})

	log.log(LevelInfo, "[INF]", "auth", "token", "secret", "agent", "codex")

	AssertContains(t, out.String(), `token="[REDACTED]"`)
	AssertNotContains(t, out.String(), "secret")
}
func TestLog_Log_shouldLog_Good(t *T) {
	log := NewLog(LogOptions{Level: LevelInfo})

	AssertTrue(t, log.shouldLog(LevelWarn))
}
func TestLog_Log_shouldLog_Bad(t *T) {
	log := NewLog(LogOptions{Level: LevelWarn})

	AssertFalse(t, log.shouldLog(LevelDebug))
}
func TestLog_Log_shouldLog_Ugly(t *T) {
	log := NewLog(LogOptions{Level: LevelQuiet})

	AssertFalse(t, log.shouldLog(LevelError))
}
func TestLog_identity_Good(t *T) {
	AssertEqual(t, "agent", identity("agent"))
}
func TestLog_identity_Bad(t *T) {
	AssertEqual(t, "", identity(""))
}
func TestLog_identity_Ugly(t *T) {
	AssertEqual(t, "colour", identity("colour"))
}
func TestLog_init_Good(t *T) {
	AssertNotNil(t, Default())
}
func TestLog_init_Bad(t *T) {
	RequireTrue(t, Default() != nil)

	AssertTrue(t, Default().shouldLog(LevelError))
}
func TestLog_init_Ugly(t *T) {
	RequireTrue(t, Default() != nil)

	AssertFalse(t, Default().shouldLog(LevelDebug))
}
