/* global $, csrf_token */
(function () {
  // ---- helpers ----
  function escapeHtml(s){return s.replace(/[&<>"']/g,m=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[m]));}
  function authHeaders(){
    const h = {};
    const u = window.user || {};
    if (u.api_key) h.Authorization = "Bearer " + u.api_key;
    if (window.csrf_token) h["X-CSRF-Token"] = csrf_token;
    return h;
  }
  $.ajaxSetup({ headers: authHeaders() });

  function api(url, opts){
    const o = Object.assign({ method:"GET", dataType:"json" }, opts||{});
    o.url = url;
    o.headers = Object.assign({}, authHeaders(), o.headers||{});
    return $.ajax(o);
  }

  function isE164(s){ return /^\+[1-9]\d{6,14}$/.test(s); }  // PT/CV included

  // ---- table render ----
  function row(p){
    return `<tr data-id="${p.id}" data-account_sid="${escapeHtml(p.account_sid)}" data-name="${escapeHtml(p.name)}"
              data-from="${escapeHtml(p.from_number)}" data-rate="${p.rate_limit_per_min}">
      <td>${p.id}</td>
      <td>${escapeHtml(p.name)}</td>
      <td>${escapeHtml(p.from_number)}</td>
      <td>${p.rate_limit_per_min}</td>
      <td>
        <button class="btn btn-xs btn-default edit"><i class="fa fa-pencil"></i> Edit</button>
        <button class="btn btn-xs btn-danger delete"><i class="fa fa-trash"></i> Delete</button>
      </td></tr>`;
  }

  function render(items){
    const $tb = $("#tbl-profiles tbody").empty();
    if (!items || items.length===0){ $("#empty-state").show(); return; }
    $("#empty-state").hide();
    items.forEach(p => $tb.append(row(p)));
  }

  function list(){
    api("/api/sms/profiles/", { method:"GET" })
      .done(render)
      .fail(xhr=>alert("List failed: "+xhr.status+" "+(xhr.responseText||"")));
  }

  // ---- modal lifecycle ----
  $("#modal-profile").on("show.bs.modal", function (ev){
    const $m = $(this);
    const $btn = $(ev.relatedTarget);
    const isEdit = $btn && $btn.hasClass("edit");

    $m.find(".modal-title").text("Add SMS Profile");
    $m.find("input[name=id]").val("");
    $m.find("input[name=name]").val("");
    $m.find("input[name=account_sid]").val("");
    $m.find("input[name=auth_token]").val("");
    $m.find("input[name=from_number]").val("");
    $m.find("input[name=rate_limit_per_min]").val("60");

    if (isEdit){
      const $tr = $btn.closest("tr");
      $m.find(".modal-title").text("Edit SMS Profile #"+$tr.data("id"));
      $m.find("input[name=id]").val($tr.data("id"));
      $m.find("input[name=name]").val($tr.data("name"));
      $m.find("input[name=account_sid]").val($tr.data("account_sid"));
      $m.find("input[name=auth_token]").val("");
      $m.find("input[name=from_number]").val($tr.data("from"));
      $m.find("input[name=rate_limit_per_min]").val($tr.data("rate"));
    }
  });

  $("#tbl-profiles").on("click","button.edit", function(e){
    e.preventDefault();
    $("#modal-profile").modal("show", this);
  });

  // ---- create/update ----
  $("#form-profile").on("submit", function(e){
    e.preventDefault();
    const id=this.id.value.trim();
    const payload={
      name: this.name.value.trim(),
      account_sid: this.account_sid.value.trim(),
      from_number: this.from_number.value.trim(),
      rate_limit_per_min: parseInt(this.rate_limit_per_min.value||"60",10)
    };
    const tok=this.auth_token.value; if(tok) payload.auth_token=tok;

    if(!payload.name || !payload.account_sid || !isFinite(payload.rate_limit_per_min) || payload.rate_limit_per_min<1 || !isE164(payload.from_number)){
      alert("Check fields. From must be E.164 (e.g., +3519XXXXXXXX or +2389XXXXXX)."); return;
    }

    if(id){
      api("/api/sms/profiles/"+id, {
        method:"PUT",
        contentType:"application/json",
        data: JSON.stringify(payload)
      })
      .done(()=>{ $("#modal-profile").modal("hide"); list(); })
      .fail(xhr=>alert("Update failed: "+xhr.status+" "+(xhr.responseText||"")));
    } else {
      api("/api/sms/profiles/", {
        method:"POST",
        contentType:"application/json",
        data: JSON.stringify(payload)
      })
      .done(()=>{ $("#modal-profile").modal("hide"); list(); })
      .fail(xhr=>alert("Create failed: "+xhr.status+" "+(xhr.responseText||"")));
    }
  });

  // ---- delete ----
  $("#tbl-profiles").on("click","button.delete", function(){
    const id=$(this).closest("tr").data("id");
    if(!confirm("Delete profile #"+id+"?")) return;
    api("/api/sms/profiles/"+id, { method:"DELETE" })
      .done(list)
      .fail(xhr=>alert("Delete failed: "+xhr.status+" "+(xhr.responseText||"")));
  });

  // live E.164 hint
  $(document).on("input", "input[name='from_number']", function(){
    $("#e164-help").css("color", isE164(this.value.trim()) ? "#3c763d" : "#a94442");
  });

  // init
  $(list);
})();
