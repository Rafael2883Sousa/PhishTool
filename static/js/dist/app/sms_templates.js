/* global $, user, csrf_token */
(function(){
  function auth(){ const h={}; const u=window.user||{}; if(u.api_key) h.Authorization="Bearer "+u.api_key; if(window.csrf_token) h["X-CSRF-Token"]=csrf_token; return h; }
  function api(url,opt){ opt=opt||{}; opt.url=url; opt.headers=Object.assign({},auth(),opt.headers||{}); return $.ajax(opt); }
  function escapeHtml(s){return s.replace(/[&<>"']/g,m=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[m]));}
  function enc(text){ const gsm="@£$¥èéùìòÇ\nØø\rÅåΔ_ΦΓΛΩΠΨΣΘΞ ^{}\\[~]|€ !\"#¤%&'()*+,-./0123456789:;<=>?ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyzÄÖÑÜ§äöñüà"; for(let ch of text) if(gsm.indexOf(ch)<0) return "UCS-2"; return "GSM-7"; }
  function segs(text){ const e=enc(text), n=text.length; return e==="GSM-7"?{enc:e,chars:n,segments:n<=160?1:Math.ceil(n/153)}:{enc:e,chars:n,segments:n<=70?1:Math.ceil(n/67)}; }
  function updateStats(){ const t=$("textarea[name=body]").val(); const s=segs(t); $("#stats").text(`Chars: ${s.chars} • Segments: ${s.segments} • Encoding: ${s.enc}`); }
  function row(t){ const p=t.body.length>80?t.body.slice(0,77)+"...":t.body; return `<tr data-id="${t.id}" data-name="${escapeHtml(t.name)}" data-body="${escapeHtml(t.body)}"><td>${t.id}</td><td>${escapeHtml(t.name)}</td><td>${escapeHtml(p)}</td><td><button class="btn btn-xs btn-default edit">Edit</button> <button class="btn btn-xs btn-danger delete">Delete</button></td></tr>`; }
  function render(items){ const $tb=$("#tbl-tpl tbody").empty(); if(!items||!items.length){$("#empty-tpl").show();return} $("#empty-tpl").hide(); items.forEach(x=>$tb.append(row(x))); }
  function list(){ api("/api/sms/templates/",{method:"GET",dataType:"json"}).done(render).fail(x=>alert("List failed: "+x.status+" "+(x.responseText||""))); }
  $("#btn-new").on("click",()=>{ const $m=$("#modal-tpl"); $m.find(".modal-title").text("Add Template"); $m.find("input[name=id]").val(""); $m.find("input[name=name]").val(""); $m.find("textarea[name=body]").val(""); updateStats(); });
  $("#tbl-tpl").on("click","button.edit",function(){ const $tr=$(this).closest("tr"), $m=$("#modal-tpl"); $m.modal("show"); $m.find(".modal-title").text("Edit Template #"+$tr.data("id")); $m.find("input[name=id]").val($tr.data("id")); $m.find("input[name=name]").val($tr.data("name")); $m.find("textarea[name=body]").val($tr.data("body")); updateStats(); });
  $("#form-tpl textarea[name='body']").on("input", updateStats);
  $("#form-tpl").on("submit",function(e){ e.preventDefault(); const id=this.id.value.trim(); const payload={name:this.name.value.trim(), body:this.body.value}; if(!payload.name||!payload.body){alert("Fill all fields");return}
    if(id){ api("/api/sms/templates/"+id,{method:"PUT",contentType:"application/json",data:JSON.stringify(payload)}).done(()=>{$("#modal-tpl").modal("hide"); list()}).fail(x=>alert("Update failed: "+x.status+" "+(x.responseText||""))); }
    else { api("/api/sms/templates/",{method:"POST",contentType:"application/json",data:JSON.stringify(payload)}).done(()=>{$("#modal-tpl").modal("hide"); list()}).fail(x=>alert("Create failed: "+x.status+" "+(x.responseText||""))); }
  });
  $(function(){ updateStats(); list(); });
})();
