/* global $, csrf_token */
(function () {
  function auth(){ const h={}; const u=window.user||{}; if(u&&u.api_key) h.Authorization="Bearer "+u.api_key; if(window.csrf_token) h["X-CSRF-Token"]=csrf_token; return h; }
  function api(url,opt){ opt=opt||{}; opt.url=url; opt.headers=Object.assign({},auth(),opt.headers||{}); return $.ajax(opt); }
  function isE164(s){ return /^\+[1-9]\d{6,14}$/.test(s); }
  function esc(s){ return String(s||"").replace(/[&<>"']/g,m=>({ "&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;" }[m])); }

  let dt;
  function render(rows){
    const $tb=$("#sms-profiles-table tbody").empty();
    rows.forEach(p=>{
      $tb.append(
        `<tr data-id="${p.id}" data-name="${esc(p.name)}" data-from="${esc(p.from_number)}"
             data-rate="${p.rate_limit_per_min}" data-account_sid="${esc(p.account_sid)}">
           <td>${esc(p.name)}</td>
           <td>${esc(p.from_number)}</td>
           <td>${p.rate_limit_per_min}</td>
           <td>
             <button class="btn btn-xs btn-default edit"><i class="fa fa-pencil"></i></button>
             <button class="btn btn-xs btn-danger delete"><i class="fa fa-trash"></i></button>
           </td>
         </tr>`);
    });
    if (!dt) dt=$("#sms-profiles-table").DataTable();
    else dt.rows().invalidate().draw();
  }
  function list(){ api("/api/sms/profiles/",{method:"GET",dataType:"json"}).done(render).fail(x=>alert("List failed: "+x.status)); }

  $("#btn-new").on("click", function(){
    const $m=$("#modal-profile");
    $m.find(".modal-title").text("Add SMS Profile");
    $m.find("input[name=id]").val("");
    $m.find("input[name=name]").val("");
    $m.find("input[name=account_sid]").val("");
    $m.find("input[name=auth_token]").val("");
    $m.find("input[name=from_number]").val("");
    $m.find("input[name=rate_limit_per_min]").val("60");
  });

  $("#sms-profiles-table").on("click",".edit", function(){
    const $tr=$(this).closest("tr"), $m=$("#modal-profile");
    $m.find(".modal-title").text("Edit SMS Profile");
    $m.find("input[name=id]").val($tr.data("id"));
    $m.find("input[name=name]").val($tr.data("name"));
    $m.find("input[name=account_sid]").val($tr.data("account_sid"));
    $m.find("input[name=auth_token]").val("");
    $m.find("input[name=from_number]").val($tr.data("from"));
    $m.find("input[name=rate_limit_per_min]").val($tr.data("rate"));
    $m.modal("show");
  });

  $("#sms-profiles-table").on("click",".delete", function(){
    const id=$(this).closest("tr").data("id");
    if(!confirm("Delete profile #"+id+"?")) return;
    api("/api/sms/profiles/"+id,{method:"DELETE"}).done(list).fail(x=>alert("Delete failed: "+x.status));
  });

  $("#form-profile").on("submit", function(e){
    e.preventDefault();
    const f=this;
    const id=(f.id.value||"").trim();
    const payload={
      name:(f.name.value||"").trim(),
      account_sid:(f.account_sid.value||"").trim(),
      from_number:(f.from_number.value||"").trim(),
      rate_limit_per_min: parseInt(f.rate_limit_per_min.value||"60",10)
    };
    const tok=(f.auth_token.value||"").trim(); if(tok) payload.auth_token=tok;
    if(!payload.name || !payload.account_sid || !isE164(payload.from_number) || payload.rate_limit_per_min<1){ alert("Invalid fields."); return; }
    const req = id
      ? { url:"/api/sms/profiles/"+id, method:"PUT", contentType:"application/json", data:JSON.stringify(payload) }
      : { url:"/api/sms/profiles/",    method:"POST", contentType:"application/json", data:JSON.stringify(payload) };
    $.ajax(Object.assign(req,{ headers:auth() }))
      .done(()=>{ $("#modal-profile").modal("hide"); list(); })
      .fail(x=>alert((id?"Update":"Create")+" failed: "+x.status+" "+(x.responseText||"")));
  });

  $(list);
})();
