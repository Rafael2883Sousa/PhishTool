/* global $, csrf_token */
(function () {
  function escapeHtml(s){return s.replace(/[&<>"']/g,m=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[m]));}
  function authHeaders(){
    const h = {};
    const u = window.user || {}; // evita ReferenceError se .User não vier do template
    if (u.api_key) h.Authorization = "Bearer " + u.api_key;
    if (window.csrf_token) h["X-CSRF-Token"] = csrf_token;
    return h;
  }
  function api(url,opts){
    const o = Object.assign({method:"GET"}, opts||{});
    o.url = url;
    o.headers = Object.assign({}, authHeaders(), o.headers||{});
    return $.ajax(o);
  }
  function isE164(s){ return /^\+[1-9]\d{6,14}$/.test(s); } // cobre PT e CV

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
    api("/api/sms/profiles/", { dataType:"json" })
      .done(render)
      .fail(xhr=>alert("List failed: "+xhr.status+" "+(xhr.responseText||"")));
  }

  // Preparação do modal quando abre via data-toggle
  $("#modal-profile").on("show.bs.modal", function (ev){
    const $m = $(this);
    const $btn = $(ev.relatedTarget);
    const isEdit = $btn && $btn.hasClass("edit"); // se abriu a partir do botão Edit (fallback)

    // Por omissão: criar
    $m.find(".modal-title").text("Add SMS Profile");
    $m.find("input[name=id]").val("");
    $m.find("input[name=name]").val("");
    $m.find("input[name=account_sid]").val("");
    $m.find("input[name=auth_token]").val("");
    $m.find("input[name=from_number]").val("");
    $m.find("input[name=rate_limit_per_min]").val("60");

    // Se foi aberto a partir de uma linha com .edit, preencher para edição
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

  // Abrir modal em modo edição (continua a funcionar sem data-target)
  $("#tbl-profiles").on("click","button.edit", function(e){
    e.preventDefault();
    $("#modal-profile").modal("show");
  });

  // Submit modal (create/update)
  $("#form-profile").on("submit", function(e){
    e.preventDefault();
    const id=this.id.value.trim(), name=this.name.value.trim(),
          sid=this.account_sid.value.trim(), tok=this.auth_token.value.trim(),
          from=this.from_number.value.trim(), rate=parseInt(this.rate_limit_per_min.value||"60",10);

    if(!name || !sid || !isFinite(rate) || rate<1 || !isE164(from)){
      alert("Check fields. 'From Number' must be E.164 (e.g., +351XXXXXXXXX or +238XXXXXXX)."); return;
    }
    const payload={ name, account_sid:sid, from_number:from, rate_limit_per_min:rate };
    if(tok) payload.auth_token=tok;

    if(id){
      api("/api/sms/profiles/"+id, { method:"PUT", contentType:"application/json", data: JSON.stringify(payload)})
        .done(()=>{ $("#modal-profile").modal("hide"); list(); })
        .fail(xhr=>alert("Update failed: "+xhr.status+" "+(xhr.responseText||"")));
    }else{
      api("/api/sms/profiles/", { method:"POST", contentType:"application/json", data: JSON.stringify(Object.assign({auth_token:tok}, payload))})
        .done(()=>{ $("#modal-profile").modal("hide"); list(); })
        .fail(xhr=>alert("Create failed: "+xhr.status+" "+(xhr.responseText||"")));
    }
  });

  // Delete
  $("#tbl-profiles").on("click","button.delete", function(){
    const id=$(this).closest("tr").data("id");
    if(!confirm("Delete profile #"+id+"?")) return;
    api("/api/sms/profiles/"+id, { method:"DELETE" })
      .done(list)
      .fail(xhr=>alert("Delete failed: "+xhr.status+" "+(xhr.responseText||"")));
  });

  // validação visual do E.164 enquanto escreve
  $(document).on("input", "input[name='from_number']", function(){
    $("#e164-help").css("color", isE164(this.value.trim()) ? "#3c763d" : "#a94442");
  });

  $(list);
})();
