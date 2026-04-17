# Disable Comment Test

<!-- gomarklint-disable -->
https://block-all-suppressed.example.com
<!-- gomarklint-enable -->
https://after-enable-all.example.com

<!-- gomarklint-disable no-bare-urls -->
https://block-named-suppressed.example.com
<!-- gomarklint-enable no-bare-urls -->
https://after-enable-named.example.com

https://disable-line-all.example.com <!-- gomarklint-disable-line -->

https://disable-line-named.example.com <!-- gomarklint-disable-line no-bare-urls -->

<!-- gomarklint-disable-next-line -->
https://next-line-all-suppressed.example.com

<!-- gomarklint-disable-next-line no-bare-urls -->
https://next-line-named-suppressed.example.com

https://wrong-rule-name.example.com <!-- gomarklint-disable-line no-bare-url -->

<!-- gomarklint-disable-next-line nonexistent-rule -->
https://nonexistent-rule.example.com
