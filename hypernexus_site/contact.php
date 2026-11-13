<?php
header('Content-Type: text/html; charset=utf-8');
$to = 'enterprise@hypernexus.site';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $name    = htmlspecialchars($_POST['name'] ?? '', ENT_QUOTES);
    $email   = htmlspecialchars($_POST['email'] ?? '', ENT_QUOTES);
    $company = htmlspecialchars($_POST['company'] ?? '', ENT_QUOTES);
    $message = htmlspecialchars($_POST['message'] ?? '', ENT_QUOTES);
    $seats   = htmlspecialchars($_POST['seats'] ?? '', ENT_QUOTES);

    $subject = "HyperNexus Demo Request - $company";
    $headers = "From: noreply@hypernexus.site\r\nReply-To: $email\r\nContent-Type: text/plain; charset=utf-8\r\nCc: $to\r\n";
    $body    = "Name: $name\nEmail: $email\nCompany: $company\nTeam Size: $seats\n\nMessage:\n$message";

    if (mail($to, $subject, $body, $headers)) {
        $success = true;
    } else {
        $error = 'Failed to send. Please try again or email us directly.';
    }
}
?>
<!DOCTYPE html><html lang="en"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Contact — HyperNexus Enterprise</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&display=swap');
body{background:#03050a;color:#f0f4f8;font-family:'Inter',sans-serif;line-height:1.6;min-height:100vh;display:flex;flex-direction:column;align-items:center;justify-content:center;padding:40px 20px}
.container{max-width:560px;width:100%;background:rgba(255,255,255,0.015);border:1px solid rgba(255,255,255,0.06);border-radius:16px;padding:48px 40px;backdrop-filter:blur(8px)}
h1{font-size:28px;font-weight:800;margin-bottom:4px;background:linear-gradient(135deg,#4d7cff,#7c5cff,#b45cff);-webkit-background-clip:text;-webkit-text-fill-color:transparent;background-clip:text}
p{color:#8899b0;font-size:14px;margin-bottom:28px}
.form-group{margin-bottom:20px}
label{display:block;font-size:13px;font-weight:600;color:#f0f4f8;margin-bottom:6px}
input,textarea{width:100%;padding:12px 14px;background:rgba(255,255,255,0.02);border:1px solid rgba(255,255,255,0.08);border-radius:8px;color:#f0f4f8;font-size:14px;font-family:'Inter',sans-serif;transition:border-color 0.3s}
input:focus,textarea:focus{outline:none;border-color:rgba(77,124,255,0.4);box-shadow:0 0 20px rgba(77,124,255,0.1)}
textarea{min-height:120px;resize:vertical}
button{width:100%;padding:14px;background:linear-gradient(135deg,#4d7cff,#7c5cff,#b45cff);color:white;border:none;border-radius:8px;font-size:15px;font-weight:600;cursor:pointer;transition:all 0.3s;box-shadow:0 0 30px rgba(77,124,255,0.2)}
button:hover{transform:translateY(-2px);box-shadow:0 0 50px rgba(77,124,255,0.35)}
.success{background:rgba(0,229,160,0.08);border:1px solid rgba(0,229,160,0.15);border-radius:8px;padding:16px;color:#00e5a0;text-align:center;margin-bottom:20px}
.error{background:rgba(255,77,77,0.08);border:1px solid rgba(255,77,77,0.15);border-radius:8px;padding:16px;color:#ff4d4d;text-align:center;margin-bottom:20px}
.back{display:block;text-align:center;margin-top:24px;color:#8899b0;text-decoration:none;font-size:13px;transition:color 0.2s}
.back:hover{color:#f0f4f8}
</style></head><body>
<div class="container">
<?php if (isset($success) && $success): ?>
<div class="success">Thank you! Your demo request has been sent. We'll be in touch within 24 hours.</div>
<a href="/" class="back">Back to HyperNexus</a>
<?php else: ?>
<h1>Request a Demo</h1>
<p>Fill out the form and our team will reach out within 24 hours.</p>
<?php if (isset($error)): ?><div class="error"><?= $error ?></div><?php endif; ?>
<form method="post" action="contact.php">
<div class="form-group"><label>Name</label><input type="text" name="name" required></div>
<div class="form-group"><label>Email</label><input type="email" name="email" required></div>
<div class="form-group"><label>Company</label><input type="text" name="company" required></div>
<div class="form-group"><label>Team Size</label><input type="text" name="seats" placeholder="e.g. 10-50"></div>
<div class="form-group"><label>Message</label><textarea name="message" required placeholder="Tell us about your AI orchestration needs..."></textarea></div>
<button type="submit">Send Demo Request</button>
</form>
<a href="/" class="back">Back to HyperNexus</a>
<?php endif; ?>
</div>
</body></html>
