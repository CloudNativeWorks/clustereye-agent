//go:build !windows
// +build !windows

package postgres

import (
	"log"
	"os/exec"
	"strings"
	"syscall"
)

// GetDiskUsage Unix/Linux sistemleri için disk kullanım bilgilerini alır
func GetDiskUsage() (string, int, string, string, string) {
	// PostgreSQL veri dizinini bul
	dataDir, err := getDataDirectoryFromConfig()
	log.Printf("DEBUG: PostgreSQL GetDiskUsage - Data directory: '%s', error: %v", dataDir, err)
	if err != nil {
		log.Printf("PostgreSQL veri dizini bulunamadı: %v", err)
		return "N/A", 0, "N/A", "N/A", "N/A"
	}

	// Disk kullanım bilgilerini al
	var stat syscall.Statfs_t
	err = syscall.Statfs(dataDir, &stat)
	log.Printf("DEBUG: PostgreSQL GetDiskUsage - Statfs for '%s': error=%v", dataDir, err)
	if err != nil {
		log.Printf("Disk kullanım bilgileri alınamadı: %v", err)
		return "N/A", 0, "N/A", "N/A", "N/A"
	}

	// Boş alanı hesapla
	freeBytes := stat.Bfree * uint64(stat.Bsize)
	totalBytes := stat.Blocks * uint64(stat.Bsize)
	usedBytes := totalBytes - freeBytes

	// Kullanım yüzdesini hesapla (Used Disk Percent)
	percent := int((float64(usedBytes) / float64(totalBytes)) * 100)

	// Boş alanı okunabilir formata çevir
	freeDisk := convertSize(freeBytes)
	totalDisk := convertSize(totalBytes)

	// Filesystem ve mount point bilgilerini al
	filesystem, mountPoint := getFilesystemInfo(dataDir)

	log.Printf("DEBUG: PostgreSQL GetDiskUsage - FINAL RESULT: freeDisk=%s, percent=%d, totalDisk=%s, filesystem=%s, mountPoint=%s",
		freeDisk, percent, totalDisk, filesystem, mountPoint)

	return freeDisk, percent, totalDisk, filesystem, mountPoint
}

// getFilesystemInfo veri dizini için filesystem ve mount point bilgilerini döndürür
func getFilesystemInfo(dataDir string) (string, string) {
	// df komutunu çalıştırarak filesystem bilgilerini al
	cmd := exec.Command("df", "-h", dataDir)
	out, err := cmd.Output()
	if err != nil {
		log.Printf("Filesystem bilgileri alınamadı: %v", err)
		return "N/A", "N/A"
	}

	// Çıktıyı satırlara böl
	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return "N/A", "N/A"
	}

	// İkinci satırı işle (ilk satır başlık)
	fields := strings.Fields(lines[1])
	if len(fields) < 6 {
		return "N/A", "N/A"
	}

	filesystem := fields[0]
	mountPoint := fields[5]

	return filesystem, mountPoint
}
