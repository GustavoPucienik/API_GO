package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	mw "systemapi/internal/middleware"
	"systemapi/pkg/response"
)

type Handler struct {
	auth        *AuthService
	user        *UserService
	group       *GroupService
	userGroup   *UserGroupService
	accessLevel *AccessLevelService
	module      *ModuleService
	menu        *MenuService
}

// ─── Auth ─────────────────────────────────────────────────────────────────────

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	data, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		response.Err(w, http.StatusUnauthorized, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	user := mw.GetUser(r)
	data, err := h.auth.RefreshUser(r.Context(), user.ID)
	if err != nil {
		response.Err(w, http.StatusUnauthorized, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) registerSelf(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	user, err := h.user.Create(r.Context(), req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Created(w, user)
}

// ─── User ─────────────────────────────────────────────────────────────────────

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	user, err := h.user.Create(r.Context(), req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Created(w, user)
}

func (h *Handler) findAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.user.FindAll(r.Context())
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, users)
}

func (h *Handler) findUsersPaginated(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	search := r.URL.Query().Get("search")

	users, total, err := h.user.FindPaginated(r.Context(), page, limit, search)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}
	response.Page(w, users, response.Meta{Total: total, Page: page, TotalPages: totalPages, Limit: limit})
}

func (h *Handler) findUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	user, err := h.user.FindByID(r.Context(), id)
	if err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Ok(w, user)
}

func (h *Handler) getUserWithRelations(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	user, err := h.user.GetWithRelations(r.Context(), id)
	if err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Ok(w, user)
}

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	user, err := h.user.Update(r.Context(), id, req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Ok(w, user)
}

func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	if err := h.user.Delete(r.Context(), id); err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Msg(w, "usuário excluído com sucesso")
}

func (h *Handler) resetPasswordAdmin(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req ResetPasswordAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if err := h.user.ResetPasswordAdmin(r.Context(), id, req); err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Msg(w, "senha alterada com sucesso")
}

func (h *Handler) resetPasswordSelf(w http.ResponseWriter, r *http.Request) {
	user := mw.GetUser(r)
	var req ResetPasswordSelfRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if err := h.user.ResetPasswordSelf(r.Context(), user.ID, req); err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Msg(w, "senha alterada com sucesso")
}

func (h *Handler) findUsersByModule(w http.ResponseWriter, r *http.Request) {
	module := r.URL.Query().Get("module")
	if module == "" {
		response.Err(w, http.StatusBadRequest, "parâmetro 'module' é obrigatório")
		return
	}

	if r.URL.Query().Get("assignable") == "true" {
		users, err := h.user.FindAssignableByModule(r.Context(), module)
		if err != nil {
			response.Err(w, http.StatusInternalServerError, err.Error())
			return
		}
		response.Ok(w, users)
		return
	}

	scope := r.URL.Query().Get("scope")
	if scope == "" {
		scope = "ALL"
	}
	users, err := h.user.FindByModuleAndScope(r.Context(), module, scope)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, users)
}

// ─── Group ────────────────────────────────────────────────────────────────────

func (h *Handler) createGroup(w http.ResponseWriter, r *http.Request) {
	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	group, err := h.group.Create(r.Context(), req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Created(w, group)
}

func (h *Handler) findGroupsPaginated(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	search := r.URL.Query().Get("search")

	groups, total, err := h.group.FindPaginated(r.Context(), page, limit, search)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}
	response.Page(w, groups, response.Meta{Total: total, Page: page, TotalPages: totalPages, Limit: limit})
}

func (h *Handler) findGroupByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	group, err := h.group.FindByID(r.Context(), id)
	if err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Ok(w, group)
}

func (h *Handler) updateGroup(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	group, err := h.group.Update(r.Context(), id, req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Ok(w, group)
}

func (h *Handler) deleteGroup(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	if err := h.group.Delete(r.Context(), id); err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Msg(w, "grupo excluído com sucesso")
}

// ─── Access Level ─────────────────────────────────────────────────────────────

func (h *Handler) createAccessLevel(w http.ResponseWriter, r *http.Request) {
	var req CreateAccessLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	al, err := h.accessLevel.Create(r.Context(), req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Created(w, al)
}

func (h *Handler) findAccessLevelsPaginated(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	search := r.URL.Query().Get("search")

	items, total, err := h.accessLevel.FindPaginated(r.Context(), page, limit, search)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}
	response.Page(w, items, response.Meta{Total: total, Page: page, TotalPages: totalPages, Limit: limit})
}

func (h *Handler) findAccessLevelByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	al, err := h.accessLevel.FindByID(r.Context(), id)
	if err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Ok(w, al)
}

func (h *Handler) updateAccessLevel(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req UpdateAccessLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	al, err := h.accessLevel.Update(r.Context(), id, req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Ok(w, al)
}

func (h *Handler) deleteAccessLevel(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	if err := h.accessLevel.Delete(r.Context(), id); err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Msg(w, "nível de acesso deletado com sucesso")
}

// ─── Module ───────────────────────────────────────────────────────────────────

func (h *Handler) createModule(w http.ResponseWriter, r *http.Request) {
	var req CreateModuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	m, err := h.module.Create(r.Context(), req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Created(w, m)
}

func (h *Handler) findAllModules(w http.ResponseWriter, r *http.Request) {
	modules, err := h.module.FindAll(r.Context())
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, modules)
}

func (h *Handler) findModuleByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	m, err := h.module.FindByID(r.Context(), id)
	if err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Ok(w, m)
}

func (h *Handler) updateModule(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req UpdateModuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	m, err := h.module.Update(r.Context(), id, req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Ok(w, m)
}

func (h *Handler) deleteModule(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	if err := h.module.Delete(r.Context(), id); err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Msg(w, "módulo excluído com sucesso")
}

// ─── Menu ─────────────────────────────────────────────────────────────────────

func (h *Handler) createMenu(w http.ResponseWriter, r *http.Request) {
	var req CreateMenuRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	m, err := h.menu.Create(r.Context(), req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Created(w, m)
}

func (h *Handler) findAllMenus(w http.ResponseWriter, r *http.Request) {
	menus, err := h.menu.FindAll(r.Context())
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, menus)
}

func (h *Handler) updateMenu(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req UpdateMenuRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	m, err := h.menu.Update(r.Context(), id, req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Ok(w, m)
}

func (h *Handler) deleteMenu(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	if err := h.menu.Delete(r.Context(), id); err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Msg(w, "menu excluído com sucesso")
}

// ─── User Group ──────────────────────────────────────────────────────────────

func (h *Handler) listUsersByGroup(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.ParseInt(chi.URLParam(r, "groupId"), 10, 32)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "groupId inválido")
		return
	}
	users, err := h.userGroup.FindUsersByGroup(r.Context(), int32(groupID))
	if err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Ok(w, users)
}

func (h *Handler) syncUsersInGroup(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.ParseInt(chi.URLParam(r, "groupId"), 10, 32)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "groupId inválido")
		return
	}
	var req SyncUsersInGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if err := h.userGroup.SyncUsersInGroup(r.Context(), int32(groupID), req.UserIDs); err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Msg(w, "usuários sincronizados com sucesso")
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}
