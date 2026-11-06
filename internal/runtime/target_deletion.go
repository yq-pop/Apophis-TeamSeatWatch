package runtime

import (
	"net/http"

	"github.com/google/uuid"
)

type deleteTargetAccountsRequest struct {
	AccountIDs []uuid.UUID `json:"accountIds"`
	Confirmed  bool        `json:"confirmed"`
}

type deleteTargetAccountsResponse struct {
	Deleted int `json:"deleted"`
}

func (h *OwnerAuthHandler) deleteTargetAccounts(w http.ResponseWriter, r *http.Request) {
	owner, ok := h.authenticated(w, r, true)
	if !ok {
		return
	}
	var request deleteTargetAccountsRequest
	if !decodeJSON(w, r, &request) || !request.Confirmed || len(request.AccountIDs) == 0 || len(request.AccountIDs) > 10000 {
		writeProblem(w, r, http.StatusUnprocessableEntity, "invalid_account_deletion", "Invalid Request", "Confirm target account deletion", 0)
		return
	}
	seen := make(map[uuid.UUID]struct{}, len(request.AccountIDs))
	ids := make([]uuid.UUID, 0, len(request.AccountIDs))
	for _, id := range request.AccountIDs {
		if id == uuid.Nil {
			writeProblem(w, r, http.StatusUnprocessableEntity, "invalid_account_deletion", "Invalid Request", "Invalid target account ID", 0)
			return
		}
		if _, dup := seen[id]; !dup {
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		h.targetFailure(w, r, err)
		return
	}
	defer tx.Rollback(r.Context())

	if err = h.lockCurrentOwnerSession(r.Context(), tx, owner); err != nil {
		writeProblem(w, r, http.StatusUnauthorized, "session_expired", "Session Expired", "Authentication is required", 0)
		return
	}

	cleanupQueries := []string{
		`DELETE FROM tsw_tasks WHERE target_account_id = ANY($1::uuid[])`,
		`DELETE FROM tsw_batch_targets WHERE target_account_id = ANY($1::uuid[])`,
		`DELETE FROM tsw_standby_child_memberships WHERE target_account_id = ANY($1::uuid[])`,
		`DELETE FROM tsw_standby_child_history WHERE target_account_id = ANY($1::uuid[])`,
		`DELETE FROM tsw_operation_targets WHERE target_account_id = ANY($1::uuid[])`,
		`DELETE FROM tsw_batch_memberships WHERE target_account_id = ANY($1::uuid[])`,
		`DELETE FROM tsw_batch_zip_members WHERE target_account_id = ANY($1::uuid[])`,
		`DELETE FROM tsw_batch_zip_protections WHERE target_account_id = ANY($1::uuid[])`,
		`DELETE FROM tsw_rotation_candidate_join_intents WHERE candidate_account_id = ANY($1::uuid[])`,
		`DELETE FROM tsw_rotation_global_protections WHERE target_account_id = ANY($1::uuid[])`,
		`DELETE FROM tsw_rotation_join_usage_evidence WHERE target_account_id = ANY($1::uuid[])`,
		`DELETE FROM tsw_rotation_removal_slots WHERE original_account_id = ANY($1::uuid[]) OR candidate_account_id = ANY($1::uuid[])`,
		`DELETE FROM tsw_rotation_usage_ledger WHERE target_account_id = ANY($1::uuid[])`,
		`DELETE FROM tsw_target_credentials WHERE target_account_id = ANY($1::uuid[])`,
	}
	for _, q := range cleanupQueries {
		if _, err = tx.Exec(r.Context(), q, ids); err != nil {
			h.targetFailure(w, r, err)
			return
		}
	}
	tag, err := tx.Exec(r.Context(), `DELETE FROM tsw_target_accounts WHERE id = ANY($1::uuid[])`, ids)
	if err != nil {
		h.targetFailure(w, r, err)
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		h.targetFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, deleteTargetAccountsResponse{Deleted: int(tag.RowsAffected())})
}
