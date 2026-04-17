# Disable Comment Test

<!-- gomarklint-disable no-bare-urls -->
https://suppressed.example.com
<!-- gomarklint-enable no-bare-urls -->

https://reported.example.com <!-- gomarklint-disable-line no-bare-urls -->

<!-- gomarklint-disable-next-line no-bare-urls -->
https://suppressed-next.example.com

https://also-reported.example.com
