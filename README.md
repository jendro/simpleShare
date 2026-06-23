# Shared JSON / Text Share

Aplikasi sederhana untuk berbagi teks, JSON, atau konten apa pun antar perangkat saat bekerja dengan banyak perangkat.

## Tujuan

- Memudahkan berbagi teks dan JSON secara real-time.
- Memfasilitasi transfer konten antar device tanpa perlu email atau chat.
- Mendukung pengiriman teks bebas, bukan hanya JSON.

## Cara pakai

1. Pastikan Anda sudah menginstall Go di komputer.
2. Buka terminal dan masuk ke folder project:
   ```bash
   cd /path/ke/sharedJson
   ```
3. Jalankan server lokal dari root paket:
   ```bash
   go run .
   ```
   > `go run main.go` tidak akan bekerja karena aplikasi terdiri dari beberapa file Go.
4. Buka browser dan kunjungi:
   ```
   http://localhost:8080
   ```
5. Setelah halaman terbuka, Anda akan otomatis masuk ke room `global`.
6. Untuk membuat room privat atau mengganti username/avatar, klik tombol `Room` atau `Guest` di pojok kanan atas.
7. Ketik atau tempel teks/JSON di textarea, lalu tekan `Enter` untuk mengirim.
8. Pesan akan muncul di bawah dan dapat disalin dengan tombol copy di kanan.

## Bagaimana ini bekerja

- Aplikasi ini menggunakan WebSocket lokal untuk mengirim pesan antar browser yang terhubung ke server.
- Pesan hanya dibagikan di dalam room yang sama.
- Room `global` adalah default dan bersifat publik di server lokal.
- Room privat dibuat dengan nama room dan password, sehingga hanya user yang memasukkan nama room dan password yang sama yang dapat melihat pesan.
- Semua data dikirim langsung melalui koneksi WebSocket antara browser dan server. Tidak ada penyimpanan permanen.

## Demo

![Demo Shared JSON](demo.png)

## Fitur

- Berbagi konten real-time melalui WebSocket.
- Berbagi file dengan preview nama file sebelum dikirim.
- Salin ke clipboard dengan fallback untuk browser yang tidak mendukung API Clipboard modern.
- Pesan terbaru ditampilkan di bagian atas.

## Catatan

Aplikasi ini cocok untuk bekerja dengan beberapa perangkat, misalnya saat berpindah dari laptop ke tablet atau dari PC kantor ke komputer rumah, dan ingin menyalin teks/JSON dengan cepat.
