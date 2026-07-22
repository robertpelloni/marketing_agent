<?php
// A/B Test Router for cloud.hypernexus.site
// 50%/50% split between dark theme (A) and blue theme (B)

// Set cookie to maintain consistency for returning visitors
if (!isset($_COOKIE['hypernexus_ab_variant'])) {
    $variant = rand(0, 1) === 0 ? 'a' : 'b';
    setcookie('hypernexus_ab_variant', $variant, time() + (86400 * 30), '/'); // 30 days
} else {
    $variant = $_COOKIE['hypernexus_ab_variant'];
}

// Track variant in analytics
echo "<script>if(typeof gtag==='function')gtag('event','ab_variant',{'variant':'$variant'});</script>";

// Serve the appropriate variant
if ($variant === 'b') {
    include('cloud_login_b.html');
} else {
    include('cloud_login.html');
}
?>
