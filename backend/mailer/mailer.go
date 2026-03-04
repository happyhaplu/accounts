package mailer

import (
	"fmt"
	"net"
	"net/smtp"
	"os"
)

// Send delivers an HTML email via the SMTP server configured in .env.
func Send(to, subject, htmlBody string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = user
	}

	addr := net.JoinHostPort(host, port)
	auth := smtp.PlainAuth("", user, pass, host)

	// Build the raw RFC-2822 message with proper CRLF line endings.
	msg := "MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"From: Outcraftly <" + from + ">\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n\r\n" +
		htmlBody

	// Use the high-level smtp.SendMail which handles EHLO, STARTTLS,
	// AUTH, DATA and QUIT in the correct order and closes cleanly.
	if err := smtp.SendMail(addr, auth, from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("smtp sendmail: %w", err)
	}
	return nil
}

// ─── Email templates ──────────────────────────────────────────────────────────

// VerifyEmailBody returns the HTML for an account-verification email.
func VerifyEmailBody(link string) string {
	return shell(
		"Verify your email address",
		"Welcome to Outcraftly! Click the button below to verify your email address and activate your account.",
		"Verify email address",
		link,
		"This link expires in 24&nbsp;hours. If you didn't create an Outcraftly account, you can safely ignore this email.",
	)
}

// PasswordResetBody returns the HTML for a password-reset email.
func PasswordResetBody(link string) string {
	return shell(
		"Reset your password",
		"We received a request to reset the password for your Outcraftly account. Click below to choose a new password.",
		"Reset my password",
		link,
		"This link expires in 1&nbsp;hour. If you didn't request a password reset, no action is needed.",
	)
}

// WorkspaceInviteBody returns the HTML for a workspace invitation email.
// wsName is the workspace name, inviterName/inviterEmail identify who sent it,
// role is "member" or "owner", and link is the full accept URL.
func WorkspaceInviteBody(wsName, inviterName, inviterEmail, role, link string) string {
	by := inviterName
	if by == "" {
		by = inviterEmail
	}
	roleLabel := "member"
	if role == "owner" {
		roleLabel = "owner"
	}
	return shell(
		"You've been invited to join "+wsName,
		by+" has invited you to collaborate on <strong>"+wsName+"</strong> as a <strong>"+roleLabel+"</strong> on Outcraftly. Click the button below to accept your invitation.",
		"Accept invitation",
		link,
		"This invitation expires in 7&nbsp;days. If you weren't expecting this, you can safely ignore it.",
	)
}

// shell is the shared HTML wrapper for all transactional emails.
func shell(heading, intro, btnText, btnURL, note string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>` + heading + `</title>
</head>
<body style="margin:0;padding:0;background:#f1f3f4;font-family:'Helvetica Neue',Helvetica,Arial,sans-serif;-webkit-font-smoothing:antialiased;">
<table width="100%" cellpadding="0" cellspacing="0" border="0" style="background:#f1f3f4;padding:48px 16px">
  <tr><td align="center">
    <table width="560" cellpadding="0" cellspacing="0" border="0"
           style="max-width:560px;width:100%;background:#ffffff;border-radius:12px;
                  overflow:hidden;box-shadow:0 2px 20px rgba(0,0,0,0.08)">

      <!-- Header -->
      <tr>
        <td style="background:linear-gradient(135deg,#1565c0 0%,#0d47a1 100%);padding:28px 40px;text-align:center">
          <span style="font-size:20px;font-weight:700;color:#ffffff;letter-spacing:-0.3px">✦ Outcraftly</span>
        </td>
      </tr>

      <!-- Body -->
      <tr>
        <td style="padding:40px 40px 28px">
          <h1 style="margin:0 0 14px;font-size:22px;font-weight:700;color:#202124;letter-spacing:-0.4px">` + heading + `</h1>
          <p  style="margin:0 0 32px;font-size:15px;color:#5f6368;line-height:1.65">` + intro + `</p>
          <!-- CTA button -->
          <table cellpadding="0" cellspacing="0" border="0">
            <tr>
              <td style="border-radius:8px;background:#1a73e8">
                <a href="` + btnURL + `"
                   style="display:inline-block;padding:13px 30px;font-size:15px;font-weight:600;
                          color:#ffffff;text-decoration:none;letter-spacing:-0.1px">` + btnText + `</a>
              </td>
            </tr>
          </table>
          <p style="margin:24px 0 0;font-size:12.5px;color:#9aa0a6;line-height:1.6">
            Or paste this link in your browser:<br>
            <a href="` + btnURL + `" style="color:#1a73e8;word-break:break-all">` + btnURL + `</a>
          </p>
        </td>
      </tr>

      <!-- Footer -->
      <tr>
        <td style="padding:20px 40px 28px;border-top:1px solid #f1f3f4">
          <p style="margin:0;font-size:12px;color:#b0b4ba;line-height:1.6">
            ` + note + `<br>
            &copy; 2026 Outcraftly, Inc.
          </p>
        </td>
      </tr>

    </table>
  </td></tr>
</table>
</body>
</html>`
}
