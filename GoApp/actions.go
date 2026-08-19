package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func (g *guiApp) loadSettingsIntoInterface() {
	port := g.settings.ProxyPort
	if port < 1 || port > 65535 {
		port = 1080
	}
	g.portEntry.SetText(strconv.Itoa(port))
	g.lanCheck.SetChecked(g.settings.ListenOnLAN)
	g.debugCheck.SetChecked(g.settings.ShowDebugLog)

	profiles, _ := listProfiles()
	selected := g.settings.SelectedProfileFile
	found := false
	for _, name := range profiles {
		if strings.EqualFold(name, selected) {
			selected = name
			found = true
			break
		}
	}
	if !found && len(profiles) > 0 {
		selected = profiles[0]
	} else if !found {
		selected = ""
	}
	g.selectedProfile = selected

	g.refreshProfileSelect()
	g.refreshConfigViewer()
	g.refreshProtocolLabel()
	g.updateHeroProfile()
	if selected != "" {
		g.settings.SelectedProfileFile = selected
		_ = g.settings.save()
		g.statusDetail.SetText("Конфигурация готова к запуску.")
	} else {
		g.statusDetail.SetText("Добавьте конфигурацию WireGuard или AmneziaWG.")
	}
	g.refreshProxyAddresses()
}

// --- Profiles ------------------------------------------------------------

// refreshProfileSelect rebuilds the dropdown's options from disk and keeps
// the current selection (clearing it if that profile was deleted).
func (g *guiApp) refreshProfileSelect() {
	profiles, err := listProfiles()
	if err != nil {
		profiles = nil
	}
	g.profileSelect.SetOptions(profiles)

	if g.selectedProfile != "" {
		found := false
		for _, name := range profiles {
			if name == g.selectedProfile {
				found = true
				break
			}
		}
		if !found {
			g.selectedProfile = ""
		}
	}

	if g.selectedProfile == "" {
		g.profileSelect.ClearSelected()
		g.deleteButton.Disable()
	} else {
		g.profileSelect.SetSelected(g.selectedProfile)
		g.deleteButton.Enable()
	}
}

// onProfileSelected fires when the user picks a different entry in the
// dropdown. Switching is refused while the proxy is running.
func (g *guiApp) onProfileSelected(name string) {
	if !g.interfaceLoaded || name == "" || name == g.selectedProfile {
		return
	}
	if g.runtime.IsRunning() {
		g.profileSelect.SetSelected(g.selectedProfile)
		g.statusDetail.SetText("Остановите прокси, чтобы сменить профиль.")
		return
	}

	g.selectedProfile = name
	g.settings.SelectedProfileFile = name
	_ = g.settings.save()
	g.deleteButton.Enable()

	g.refreshConfigViewer()
	g.refreshProtocolLabel()
	g.updateHeroProfile()
	g.statusDetail.SetText("Выбрана конфигурация: " + name)
	g.appendLog("Выбрана конфигурация: " + name)
}

func (g *guiApp) updateHeroProfile() {
	if g.heroProfile == nil {
		return
	}
	if g.selectedProfile == "" {
		g.heroProfile.Text = "Профиль не выбран"
	} else {
		g.heroProfile.Text = "Профиль: " + g.selectedProfile
	}
	g.heroProfile.Refresh()
}

func (g *guiApp) selectedProfilePath() (string, bool) {
	if g.selectedProfile == "" {
		return "", false
	}
	path, err := profilePath(g.selectedProfile)
	if err != nil {
		return "", false
	}
	return path, true
}

func (g *guiApp) refreshConfigViewer() {
	path, ok := g.selectedProfilePath()
	if !ok {
		g.configViewer.SetText("")
		return
	}
	config, err := readConfigFile(path)
	if err != nil {
		g.showError("Не удалось прочитать конфигурацию", err)
		return
	}
	if g.showSecretsChk.Checked {
		g.configViewer.SetText(config)
	} else {
		g.configViewer.SetText(maskSecrets(config))
	}
}

func (g *guiApp) refreshProtocolLabel() {
	path, ok := g.selectedProfilePath()
	if !ok {
		g.protocolLabel.SetText("Протокол: не определён")
		return
	}
	config, err := readConfigFile(path)
	if err != nil {
		g.protocolLabel.SetText("Протокол: ошибка конфигурации")
		return
	}
	g.protocolLabel.SetText("Протокол: " + detectProtocol(config))
}

func (g *guiApp) onShowSecretsChanged(bool) { g.refreshConfigViewer() }

func (g *guiApp) onImportConfiguration() {
	if g.runtime.IsRunning() {
		g.statusDetail.SetText("Остановите прокси, чтобы изменить список профилей.")
		return
	}
	pickConfigFile(g.window, func(name string, raw []byte, err error) {
		if err != nil {
			g.showError("Не удалось импортировать конфигурацию", err)
			return
		}
		if raw == nil {
			return // Cancelled.
		}

		stored, err := importProfile(name, raw)
		if err != nil {
			g.showError("Не удалось импортировать конфигурацию", err)
			return
		}

		g.selectedProfile = stored
		g.settings.SelectedProfileFile = stored
		_ = g.settings.save()
		g.refreshProfileSelect()
		g.refreshConfigViewer()
		g.refreshProtocolLabel()
		g.updateHeroProfile()
		g.statusDetail.SetText("Конфигурация импортирована. Проверьте порт и нажмите «Запустить прокси».")
		g.appendLog("Добавлена конфигурация: " + stored)
	})
}

func (g *guiApp) onDeleteSelectedProfile() {
	name := g.selectedProfile
	if name == "" {
		return
	}
	if g.runtime.IsRunning() {
		g.statusDetail.SetText("Остановите прокси, чтобы изменить список профилей.")
		return
	}
	dialog.ShowConfirm("Удалить конфигурацию",
		fmt.Sprintf("Удалить «%s»? Это действие необратимо.", name),
		func(confirmed bool) {
			if !confirmed {
				return
			}
			if err := deleteProfile(name); err != nil {
				g.showError("Не удалось удалить конфигурацию", err)
				return
			}
			g.selectedProfile = ""
			g.settings.SelectedProfileFile = ""
			_ = g.settings.save()
			g.refreshProfileSelect()
			g.refreshConfigViewer()
			g.refreshProtocolLabel()
			g.updateHeroProfile()
			g.appendLog("Удалена конфигурация: " + name)
		}, g.window)
}

// --- Connection settings ---------------------------------------------------

func (g *guiApp) onPortChanged(text string) {
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, text)
	if len(digits) > 5 {
		digits = digits[:5]
	}
	if digits != text {
		g.portEntry.SetText(digits) // Re-enters this handler once with the clean value.
		return
	}
	if !g.interfaceLoaded {
		return
	}
	g.refreshProxyAddresses()
}

func (g *guiApp) onLanChanged(bool) {
	if !g.interfaceLoaded {
		return
	}
	g.refreshProxyAddresses()
}

func (g *guiApp) proxyPort() int {
	port, err := strconv.Atoi(g.portEntry.Text)
	if err != nil || port < 1 || port > 65535 {
		return 1080
	}
	return port
}

func (g *guiApp) refreshProxyAddresses() {
	g.addressLabel.SetText(g.proxyAddressText())
	g.lanWarning.Hidden = !g.lanCheck.Checked
	if g.lanCheck.Checked {
		g.lanWarning.Show()
	} else {
		g.lanWarning.Hide()
	}
}

func (g *guiApp) proxyAddressText() string {
	port := g.proxyPort()
	if !g.lanCheck.Checked {
		return fmt.Sprintf("socks5://127.0.0.1:%d", port)
	}
	addrs := lanIPv4Addresses()
	if len(addrs) == 0 {
		return fmt.Sprintf("socks5://<LAN-IP>:%d", port)
	}
	parts := make([]string, len(addrs))
	for i, a := range addrs {
		parts[i] = fmt.Sprintf("socks5://%s:%d", a, port)
	}
	return strings.Join(parts, "   •   ")
}

func (g *guiApp) onCopyAddress() {
	g.fyneApp.Clipboard().SetContent(g.proxyAddressText())
	g.statusDetail.SetText("Адрес прокси скопирован в буфер обмена.")
}

// --- Proxy lifecycle -------------------------------------------------------

func (g *guiApp) onToggleProxy() {
	if g.runtime.IsRunning() {
		g.toggleButton.SetDisabled(true)
		g.stopStatistics()
		g.runtime.Stop()
		g.setRunningState(false)
		g.appendLog("Прокси остановлен.")
		return
	}

	path, ok := g.selectedProfilePath()
	if !ok {
		g.showError("Не удалось запустить прокси", fmt.Errorf("сначала импортируйте WireGuard-конфигурацию"))
		return
	}
	config, err := readConfigFile(path)
	if err != nil {
		g.showError("Не удалось запустить прокси", err)
		return
	}
	if err := validateConfig(config); err != nil {
		g.showError("Не удалось запустить прокси", err)
		return
	}

	port := g.proxyPort()
	listenOnLAN := g.lanCheck.Checked
	if !tcpPortAvailable(port, listenOnLAN) {
		g.showError("Не удалось запустить прокси", fmt.Errorf("TCP-порт %d уже занят другой программой", port))
		return
	}

	g.settings.ProxyPort = port
	g.settings.ListenOnLAN = listenOnLAN
	_ = g.settings.save()

	proxyConfigPath, err := writeProxyConfig(path, port, listenOnLAN)
	if err != nil {
		g.showError("Не удалось подготовить конфигурацию ядра", err)
		return
	}

	g.toggleButton.SetDisabled(true)
	g.setStatusDetail("Запуск…", "Поднимаю WireGuard-туннель и открываю SOCKS5-порт.")
	g.hasHandshake = false

	if err := g.runtime.Start(proxyConfigPath, g.appendLog); err != nil {
		g.setRunningState(false)
		g.showError("Не удалось запустить прокси", err)
		return
	}

	g.setRunningState(true)
	g.startStatistics()
	g.appendLog("Прокси запущен: " + strings.ReplaceAll(g.proxyAddressText(), "   •   ", ", "))
}

func (g *guiApp) setRunningState(running bool) {
	if running {
		g.importButton.Disable()
		g.deleteButton.Disable()
		g.profileSelect.Disable()
		g.portEntry.Disable()
		g.lanCheck.Disable()
	} else {
		g.importButton.Enable()
		g.profileSelect.Enable()
		g.portEntry.Enable()
		g.lanCheck.Enable()
		if g.selectedProfile != "" {
			g.deleteButton.Enable()
		}
	}

	g.toggleButton.SetDisabled(false)
	if running {
		g.toggleButton.SetState("Остановить прокси", true)
		g.statusCapsule.SetState("Работает", true)
		g.sidebarStatusDot.SetState("Работает", true)
		g.statusDetail.SetText(strings.ReplaceAll(g.proxyAddressText(), "   •   ", "   |   "))
	} else {
		g.toggleButton.SetState("Запустить прокси", false)
		g.statusCapsule.SetState("Остановлен", false)
		g.sidebarStatusDot.SetState("Остановлен", false)
		g.statusDetail.SetText("Настройки можно изменить и запустить снова.")
		g.resetStatisticsDisplay()
	}
	g.sidebarToggle.SetChecked(running)
	g.updateTray(running)
}

func (g *guiApp) setStatusDetail(status, detail string) {
	g.statusDetail.SetText(detail)
	_ = status
}

// --- Statistics --------------------------------------------------------

const sparklineHistoryLen = 60 // one point per second, ~1 minute of trend

func (g *guiApp) startStatistics() {
	g.stopStatistics()
	g.prevReceived = 0
	g.prevSent = 0
	g.prevSampleTime = time.Now()
	g.hasPrevSample = false
	g.receivedHistory = nil
	g.sentHistory = nil
	g.peakBps = 0
	g.sessionStart = time.Now()

	ctx, cancel := context.WithCancel(context.Background())
	g.statsCancel = cancel
	go g.pollStatistics(ctx)
}

func (g *guiApp) stopStatistics() {
	if g.statsCancel != nil {
		g.statsCancel()
		g.statsCancel = nil
	}
}

func (g *guiApp) pollStatistics(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		stats, err := g.runtime.Stats()
		if err != nil {
			fyne.Do(func() {
				g.activityLabel.SetText("Статистика недоступна")
			})
			continue
		}

		now := time.Now()
		seconds := now.Sub(g.prevSampleTime).Seconds()
		if seconds < 0.1 {
			seconds = 0.1
		}
		var receivedDelta, sentDelta int64
		if g.hasPrevSample {
			if d := stats.ReceivedBytes - g.prevReceived; d > 0 {
				receivedDelta = d
			}
			if d := stats.SentBytes - g.prevSent; d > 0 {
				sentDelta = d
			}
		}
		g.prevReceived = stats.ReceivedBytes
		g.prevSent = stats.SentBytes
		g.prevSampleTime = now
		g.hasPrevSample = true

		trafficChanged := receivedDelta+sentDelta > 0
		fyne.Do(func() {
			g.updateStatisticsDisplay(stats, float64(receivedDelta)/seconds, float64(sentDelta)/seconds, trafficChanged)
		})
	}
}

func (g *guiApp) updateStatisticsDisplay(stats tunnelStats, receivedPerSecond, sentPerSecond float64, trafficChanged bool) {
	g.receivedLabel.SetText("↓ Получено: " + formatBytes(float64(stats.ReceivedBytes)))
	g.sentLabel.SetText("↑ Отправлено: " + formatBytes(float64(stats.SentBytes)))
	g.speedLabel.SetText(fmt.Sprintf("Скорость: ↓ %s/с  ↑ %s/с", formatBytes(receivedPerSecond), formatBytes(sentPerSecond)))
	if stats.HasHandshake {
		g.handshakeLabel.SetText("Handshake: " + stats.LastHandshake.Local().Format("15:04:05"))
	} else {
		g.handshakeLabel.SetText("Handshake: —")
	}

	combined := receivedPerSecond + sentPerSecond
	if combined > g.peakBps {
		g.peakBps = combined
	}
	g.peakLabel.SetText("Пик: " + formatBytes(g.peakBps) + "/с")
	if elapsed := time.Since(g.sessionStart).Seconds(); elapsed >= 1 {
		avg := float64(stats.ReceivedBytes+stats.SentBytes) / elapsed
		g.avgLabel.SetText("Средняя: " + formatBytes(avg) + "/с")
	}

	g.receivedHistory = pushSample(g.receivedHistory, receivedPerSecond, sparklineHistoryLen)
	g.sentHistory = pushSample(g.sentHistory, sentPerSecond, sparklineHistoryLen)
	g.receivedSparkline.SetSamples(g.receivedHistory)
	g.sentSparkline.SetSamples(g.sentHistory)

	switch {
	case trafficChanged:
		g.activityLabel.SetText("● Трафик идёт")
	case stats.HasHandshake && time.Since(stats.LastHandshake) < 3*time.Minute:
		g.activityLabel.SetText("● Соединение активно")
	default:
		g.activityLabel.SetText("● Нет трафика")
	}
}

func pushSample(history []float64, value float64, maxLen int) []float64 {
	history = append(history, value)
	if len(history) > maxLen {
		history = history[len(history)-maxLen:]
	}
	return history
}

func (g *guiApp) resetStatisticsDisplay() {
	g.receivedLabel.SetText("↓ Получено: 0 Б")
	g.sentLabel.SetText("↑ Отправлено: 0 Б")
	g.speedLabel.SetText("Скорость: ↓ 0 Б/с  ↑ 0 Б/с")
	g.handshakeLabel.SetText("Handshake: —")
	g.peakLabel.SetText("Пик: —")
	g.avgLabel.SetText("Средняя: —")
	g.activityLabel.SetText("Нет данных")
	g.receivedHistory = nil
	g.sentHistory = nil
	g.receivedSparkline.SetSamples(nil)
	g.sentSparkline.SetSamples(nil)
}

func formatBytes(value float64) string {
	units := []string{"Б", "КБ", "МБ", "ГБ", "ТБ"}
	unit := 0
	for value >= 1024 && unit < len(units)-1 {
		value /= 1024
		unit++
	}
	switch {
	case unit == 0:
		return fmt.Sprintf("%.0f %s", value, units[unit])
	case value >= 100:
		return fmt.Sprintf("%.0f %s", value, units[unit])
	case value >= 10:
		return fmt.Sprintf("%.1f %s", value, units[unit])
	default:
		return fmt.Sprintf("%.2f %s", value, units[unit])
	}
}

// --- Log -------------------------------------------------------------------

func (g *guiApp) onDebugChanged(checked bool) {
	g.settings.ShowDebugLog = checked
	_ = g.settings.save()
}

const maxLogRunes = 60_000

// appendLog is safe to call from any goroutine.
func (g *guiApp) appendLog(text string) {
	fyne.Do(func() {
		normalized := g.normalizeLogLine(text)
		if normalized == "" {
			return
		}
		line := "[" + time.Now().Format("15:04:05") + "] " + normalized + "\n"
		updated := g.logBox.Text + line
		if len(updated) > maxLogRunes {
			if cut := strings.IndexByte(updated[len(updated)-maxLogRunes:], '\n'); cut >= 0 {
				updated = updated[len(updated)-maxLogRunes+cut+1:]
			} else {
				updated = updated[len(updated)-maxLogRunes:]
			}
		}
		g.logBox.SetText(updated)
		g.logBox.CursorRow = strings.Count(updated, "\n")
	})
}

func (g *guiApp) normalizeLogLine(text string) string {
	lower := strings.ToLower(text)
	isHandshakeResponse := strings.Contains(lower, "received handshake response")
	isHandshakeInitiation := strings.Contains(lower, "sending handshake initiation")

	if g.debugCheck.Checked {
		if isHandshakeResponse {
			g.hasHandshake = true
		}
		return text
	}

	if isHandshakeResponse {
		if !g.hasHandshake {
			g.hasHandshake = true
			return "WireGuard-соединение установлено (handshake получен)."
		}
		return "Сеансовые ключи WireGuard обновлены."
	}

	if isHandshakeInitiation {
		if g.hasHandshake {
			return ""
		}
		return "Устанавливается WireGuard-соединение…"
	}

	isError := strings.Contains(lower, "error") || strings.Contains(lower, "failed") ||
		strings.Contains(lower, "fatal") || strings.Contains(lower, "panic")
	if isError {
		return text
	}

	if strings.HasPrefix(text, "DEBUG:") ||
		strings.Contains(lower, "resolving address for") ||
		strings.Contains(lower, "health metric request") {
		return ""
	}

	return text
}

func (g *guiApp) showError(title string, err error) {
	g.statusDetail.SetText(err.Error())
	g.appendLog("Ошибка: " + err.Error())
	dialog.ShowError(fmt.Errorf("%s: %w", title, err), g.window)
}

func readConfigFile(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return normalizeConfig(string(raw)), nil
}
