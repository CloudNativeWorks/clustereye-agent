package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// CPUStat represents CPU statistics from /proc/stat
type CPUStat struct {
	User    uint64
	Nice    uint64
	System  uint64
	Idle    uint64
	Iowait  uint64
	Irq     uint64
	Softirq uint64
	Steal   uint64
	Total   uint64
}

// GetCPUUsage returns CPU usage percentage using detailed /proc/stat method
// This is the same method used by PostgreSQL collector for consistency
func GetCPUUsage() (float64, error) {
	// Linux sistemlerde /proc/stat kullan
	if _, err := os.Stat("/proc/stat"); err == nil {
		// İlk ölçüm
		cpu1, err := readCPUStat()
		if err != nil {
			log.Printf("İlk CPU ölçümü hatası: %v", err)
			goto AlternativeMethod
		}

		// 500ms bekle (daha uzun süre ile daha doğru ölçüm)
		time.Sleep(500 * time.Millisecond)

		// İkinci ölçüm
		cpu2, err := readCPUStat()
		if err != nil {
			log.Printf("İkinci CPU ölçümü hatası: %v", err)
			goto AlternativeMethod
		}

		// Değişimleri hesapla
		userDiff := cpu2.User - cpu1.User
		niceDiff := cpu2.Nice - cpu1.Nice
		systemDiff := cpu2.System - cpu1.System
		idleDiff := cpu2.Idle - cpu1.Idle
		iowaitDiff := cpu2.Iowait - cpu1.Iowait
		irqDiff := cpu2.Irq - cpu1.Irq
		softirqDiff := cpu2.Softirq - cpu1.Softirq
		stealDiff := cpu2.Steal - cpu1.Steal
		totalDiff := cpu2.Total - cpu1.Total

		log.Printf("CPU farkları - User: %d, System: %d, Idle: %d, IOWait: %d, Total: %d",
			userDiff, systemDiff, idleDiff, iowaitDiff, totalDiff)

		if totalDiff == 0 {
			log.Printf("UYARI: Total diff 0, alternatif yönteme geçiliyor")
			goto AlternativeMethod
		}

		// CPU kullanımını hesapla (user + nice + system + irq + softirq + steal)
		activeDiff := userDiff + niceDiff + systemDiff + irqDiff + softirqDiff + stealDiff
		cpuUsage := (float64(activeDiff) / float64(totalDiff)) * 100

		// Geçerlilik kontrolü
		if cpuUsage < 0 || cpuUsage > 100 {
			log.Printf("UYARI: Geçersiz CPU kullanımı (%f), alternatif yönteme geçiliyor", cpuUsage)
			goto AlternativeMethod
		}

		log.Printf("Hesaplanan CPU kullanımı: %f", cpuUsage)
		return cpuUsage, nil
	}

AlternativeMethod:
	// Alternatif yöntem - mpstat kullan (daha doğru sonuçlar için)
	cmd := exec.Command("mpstat", "1", "1")
	out, err := cmd.Output()
	if err != nil {
		log.Printf("mpstat komutu hatası: %v, top deneniyor", err)
		// top dene
		cmd = exec.Command("sh", "-c", "top -bn2 -d 0.5 | grep '^%Cpu' | tail -1 | awk '{print 100-$8}'")
		out, err = cmd.Output()
		if err != nil {
			log.Printf("top komutu hatası: %v, vmstat deneniyor", err)
			// vmstat dene
			cmd = exec.Command("sh", "-c", "vmstat 1 2 | tail -1 | awk '{print 100-$15}'")
			out, err = cmd.Output()
			if err != nil {
				log.Printf("vmstat komutu hatası: %v", err)
				return 0, err
			}
		}
	}

	cpuPercent, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		log.Printf("CPU yüzdesi parse hatası: %v", err)
		return 0, err
	}

	// Geçerlilik kontrolü
	if cpuPercent < 0 {
		cpuPercent = 0
	} else if cpuPercent > 100 {
		cpuPercent = 100
	}

	log.Printf("Alternatif yöntem CPU kullanımı: %f", cpuPercent)
	return cpuPercent, nil
}

// readCPUStat reads CPU statistics from /proc/stat
func readCPUStat() (*CPUStat, error) {
	contents, err := os.ReadFile("/proc/stat")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == "cpu" {
			// En az 8 alan olmalı (cpu, user, nice, system, idle, iowait, irq, softirq)
			if len(fields) < 8 {
				return nil, fmt.Errorf("yetersiz CPU stat alanı")
			}

			stat := &CPUStat{}

			// Değerleri parse et
			values := make([]uint64, len(fields)-1)
			for i := 1; i < len(fields); i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					log.Printf("CPU stat parse hatası [%d]: %v", i, err)
					continue
				}
				values[i-1] = val
			}

			// Değerleri ata
			stat.User = values[0]
			stat.Nice = values[1]
			stat.System = values[2]
			stat.Idle = values[3]
			stat.Iowait = values[4]
			stat.Irq = values[5]
			stat.Softirq = values[6]
			if len(values) > 7 {
				stat.Steal = values[7]
			}

			// Toplam CPU zamanını hesapla
			stat.Total = stat.User + stat.Nice + stat.System + stat.Idle +
				stat.Iowait + stat.Irq + stat.Softirq + stat.Steal

			log.Printf("CPU stat detaylı - User: %d, System: %d, Idle: %d, IOWait: %d, Total: %d",
				stat.User, stat.System, stat.Idle, stat.Iowait, stat.Total)
			return stat, nil
		}
	}
	return nil, fmt.Errorf("CPU stats not found in /proc/stat")
}
