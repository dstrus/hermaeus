package settings

import (
	"net/http"

	"hermaeus/internal/httpapi"
)

func (s *Settings) Routes(mux *http.ServeMux) {
	mux.Handle("GET /api/settings", httpapi.JSON(func(*http.Request, httpapi.None) (FileList, error) {
		return s.List()
	}))
	mux.Handle("GET /api/settings/{name}", httpapi.JSON(func(r *http.Request, _ httpapi.None) (FileValues, error) {
		return s.Read(r.PathValue("name"))
	}))
	mux.Handle("PATCH /api/settings/{name}", httpapi.JSON(func(r *http.Request, e Edit) (FileValues, error) {
		return s.Save(r.PathValue("name"), e)
	}))
	mux.Handle("PUT /api/settings/{name}/pin", httpapi.JSON(func(r *http.Request, _ httpapi.None) (PinsResponse, error) {
		return s.SetPinned(r.PathValue("name"), true)
	}))
	mux.Handle("DELETE /api/settings/{name}/pin", httpapi.JSON(func(r *http.Request, _ httpapi.None) (PinsResponse, error) {
		return s.SetPinned(r.PathValue("name"), false)
	}))
	// An NPC's Universal Translator speech pattern, by bio template.
	mux.Handle("GET /api/speech/{template}", httpapi.JSON(func(r *http.Request, _ httpapi.None) (SpeechRule, error) {
		return s.Speech(r.PathValue("template"))
	}))
	mux.Handle("PUT /api/speech/{template}", httpapi.JSON(func(r *http.Request, e SpeechEdit) (SpeechRule, error) {
		return s.SetSpeech(r.PathValue("template"), e)
	}))
	// Universal Translator's whole configuration: its rules, filter, and switch.
	mux.Handle("GET /api/translator", httpapi.JSON(func(*http.Request, httpapi.None) (Translator, error) {
		return s.Translator()
	}))
	mux.Handle("PUT /api/translator", httpapi.JSON(func(_ *http.Request, e TranslatorEdit) (Translator, error) {
		return s.SetTranslator(e)
	}))
	mux.Handle("POST /api/translator/import", httpapi.JSON(func(_ *http.Request, imp Import) (ImportResult, error) {
		return s.ImportRules(imp)
	}))
	// Which model does each of SkyrimNet's tasks.
	mux.Handle("GET /api/models", httpapi.JSON(func(*http.Request, httpapi.None) (Models, error) {
		return s.Models()
	}))
	// A provider's own model list, fetched from it on request.
	mux.Handle("GET /api/models/providers/{id}", httpapi.JSON(func(r *http.Request, _ httpapi.None) (ModelList, error) {
		return s.ModelList(r.PathValue("id"), r.URL.Query().Has("refresh"))
	}))
	mux.Handle("PUT /api/models/{task}", httpapi.JSON(func(r *http.Request, e TaskEdit) (Models, error) {
		return s.SetTask(r.PathValue("task"), e)
	}))
	// The providers tasks can be sent to; a key is set or cleared, never read.
	mux.Handle("GET /api/providers", httpapi.JSON(func(*http.Request, httpapi.None) (ProvidersResponse, error) {
		return s.Providers()
	}))
	mux.Handle("POST /api/providers", httpapi.JSON(func(_ *http.Request, e ProviderEdit) (ProvidersResponse, error) {
		return s.AddProvider(e)
	}))
	mux.Handle("PUT /api/providers/{id}", httpapi.JSON(func(r *http.Request, e ProviderEdit) (ProvidersResponse, error) {
		return s.SetProvider(r.PathValue("id"), e)
	}))
	mux.Handle("DELETE /api/providers/{id}", httpapi.JSON(func(r *http.Request, e KeyEdit) (ProvidersResponse, error) {
		return s.RemoveProvider(r.PathValue("id"), e.Modified)
	}))
	mux.Handle("PUT /api/providers/{id}/key", httpapi.JSON(func(r *http.Request, e KeyEdit) (ProvidersResponse, error) {
		return s.SetKey(r.PathValue("id"), e)
	}))
	// Plugins' own settings, as their schemas describe them.
	mux.Handle("GET /api/plugin-settings", httpapi.JSON(func(*http.Request, httpapi.None) ([]PluginSchema, error) {
		return s.PluginSchemas(), nil
	}))
	mux.Handle("GET /api/toggles", httpapi.JSON(func(*http.Request, httpapi.None) (ToggleList, error) {
		return s.Toggles()
	}))
	mux.Handle("PUT /api/toggles/{id}", httpapi.JSON(func(r *http.Request, req ToggleRequest) (ToggleList, error) {
		return s.SetToggle(r.PathValue("id"), req.On)
	}))
}
