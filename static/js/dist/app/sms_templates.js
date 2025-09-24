/* global $, user, csrf_token */
(function(){
  function auth(){ 
    const h={}; 
    const u=window.user||{}; 
    if(u.api_key) h.Authorization="Bearer "+u.api_key; 
    if(window.csrf_token) h["X-CSRF-Token"]=csrf_token; 
    return h; 
  }

  function api(url,opt){ 
    opt=opt||{}; 
    opt.url=url; 
    opt.headers=Object.assign({},auth(),opt.headers||{}); 
    return $.ajax(opt); 
  }

  function escapeHtml(s){
    return s.replace(/[&<>"']/g,m=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[m]));
  }

  function enc(text){ 
    const gsm="@£$¥èéùìòÇ\nØø\rÅåΔ_ΦΓΛΩΠΨΣΘΞ ^{}\\[~]|€ !\"#¤%&'()*+,-./0123456789:;<=>?ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyzÄÖÑÜ§äöñüà"; 
    for(let ch of text) if(gsm.indexOf(ch)<0) return "UCS-2"; return "GSM-7"; 
  }

  function segs(text){ 
    const e=enc(text), 
    n=text.length; 
    return e==="GSM-7"?{enc:e,chars:n,segments:n<=160?1:Math.ceil(n/153)}:{enc:e,chars:n,segments:n<=70?1:Math.ceil(n/67)}; 
  }

  function updateStats(){ 
    const t=$("textarea[name=body]").val(); const s=segs(t); $("#stats").text(`Chars: ${s.chars} • Segments: ${s.segments} • Encoding: ${s.enc}`); 
  }
  
  function modalError(msg){
    $("#modal\\.flashes").empty().append(
      '<div style="text-align:center" class="alert alert-danger"><i class="fa fa-exclamation-circle"></i> '+escapeHtml(msg)+'</div>');
  }
  function successFlash(msg){
    $("#flashes").empty().append(
      '<div style="text-align:center" class="alert alert-success"><i class="fa fa-check-circle"></i> '+escapeHtml(msg)+'</div>');
    setTimeout(function(){ $("#flashes").empty(); }, 4000);
  }
  function errorFlash(msg){
    $("#flashes").empty().append(
      '<div style="text-align:center" class="alert alert-danger"><i class="fa fa-exclamation-circle"></i> '+escapeHtml(msg)+'</div>');
  }

  function reloadTable(){
    if ($.fn.dataTable && $.fn.dataTable.isDataTable('#tbl-tpl')) {
      $('#tbl-tpl').DataTable().destroy(); // evita “reinit” conflict
    }
    list();
  }


  function row(t){
    const p = t.body.length > 80 ? t.body.slice(0,77) + "..." : t.body;
    return `<tr data-id="${t.id}" data-name="${escapeHtml(t.name)}" data-body="${escapeHtml(t.body)}">
      <td>${t.id}</td>
      <td>${escapeHtml(t.name)}</td>
      <td>${escapeHtml(p)}</td>
      <td>
        <div class="pull-right">
          <button class="btn btn-primary edit" data-toggle="tooltip" data-placement="left" title="Edit Template">
            <i class="fa fa-pencil"></i>
          </button>
          <button class="btn btn-danger delete" data-toggle="tooltip" data-placement="left" title="Delete Template">
            <i class="fa fa-trash-o"></i>
          </button>
        </div>
      </td>
    </tr>`;
  }

  function initTooltips(){
    if (!$.fn || !$.fn.tooltip){ console.warn("[sms_templates] Bootstrap tooltip ausente"); return; }
    $('body').tooltip('destroy');
    $('body').tooltip({
      selector: '[data-toggle="tooltip"]',
      container: 'body',
      placement: 'left',
      trigger: 'hover'
    });
  }

  function render(items){ 

    const $tb=$("#tbl-tpl tbody").empty(); 
    if(!items||!items.length){
      $("#empty-tpl").show("#emptyMessage");
      return
    }
    $("#empty-tpl").hide(); 
    items.forEach(x=>$tb.append(row(x)));  
    initTooltips();

    // inicializa/renova DataTables com as mesmas opções do profiles
    if ($.fn.dataTable) {
      $('#tbl-tpl').DataTable({
        destroy: true,
        pageLength: 10,
        lengthMenu: [10, 25, 50, 100],
        columnDefs: [{ orderable: false, targets: 3 }],
        language: {
          lengthMenu: "Show _MENU_ entries",
          search: "Search:",
          paginate: { previous: "Previous", next: "Next" }
        }
      });
    }
  }
  
  function list(){ 
    api("/api/sms/templates/",{method:"GET",dataType:"json"}).done(render).fail(x=>errorFlash("Error fetching SMS templates ("+x.status+") "+(x.responseText||""))); 
  }

  $("#btn-new").on("click",()=>{ const $m=$("#modal-tpl"); $m.find(".modal-title").text("Add Template"); $m.find("input[name=id]").val(""); $m.find("input[name=name]").val(""); $m.find("textarea[name=body]").val(""); updateStats(); });

  $("#tbl-tpl").on("click","button.edit",function(){ 
    const $tr=$(this).closest("tr"), $m=$("#modal-tpl"); $m.modal("show"); $m.find(".modal-title").text("Edit Template #"+$tr.data("id")); $m.find("input[name=id]").val($tr.data("id")); $m.find("input[name=name]").val($tr.data("name")); $m.find("textarea[name=body]").val($tr.data("body")); updateStats(); 
  });

  $("#tbl-tpl").on("click","button.delete", function(){
    const $tr   = $(this).closest("tr");
    const id    = $tr.data("id");
    const name  = $tr.data("name");
    if (!id) return;

    if (typeof Swal === "undefined") {
      if (!confirm(`Delete template "${name}"? This can't be undone!`)) return;
      api(`/api/sms/templates/${id}`, { method:"DELETE" })
        .done(()=>{ if (typeof successFlash==="function") successFlash("Template deleted!"); reloadTable(); })
        .fail(x=>{ const m=(x.responseJSON&&x.responseJSON.message)||`Delete failed (${x.status})`; if (typeof modalError==="function") modalError(m); });
      return;
    }

    Swal.fire({
      title: "Are you sure?",
      text: "This will delete the template. This can't be undone!",
      type: "warning",
      animation: false,
      showCancelButton: true,
      confirmButtonText: "Delete " + escapeHtml(name),
      confirmButtonColor: "#428bca",
      reverseButtons: true,
      allowOutsideClick: false,
      preConfirm: () => new Promise((resolve, reject) => {
        api(`/api/sms/templates/${id}`, { method:"DELETE" })
          .done(()=>resolve())
          .fail(x=>reject((x.responseJSON&&x.responseJSON.message)||"Delete failed"));
      })
    }).then((result)=>{
      if (result.value){
        Swal.fire('Template Deleted!','This template has been deleted!','success');
      }
      $('button:contains("OK")').on('click', ()=> reloadTable());
    });
  });


  $("#form-tpl textarea[name='body']").on("input", updateStats);
  $("#form-tpl").on("submit",function(e){ e.preventDefault(); const id=this.id.value.trim(); const payload={name:this.name.value.trim(), body:this.body.value}; if(!payload.name||!payload.body){alert("Fill all fields");return}
    if(id){ api("/api/sms/templates/"+id,{method:"PUT",contentType:"application/json",data:JSON.stringify(payload)}).done(()=>{ $("#modal-tpl").one('hidden.bs.modal', reloadTable).modal('hide'); successFlash("Template edited successfully!"); list(); }).fail(x=>modalError((x.responseJSON&&x.responseJSON.message)||("Update failed ("+x.status+")"))); }
    else { api("/api/sms/templates/",{method:"POST",contentType:"application/json",data:JSON.stringify(payload)}).done(()=>{ $("#modal-tpl").one('hidden.bs.modal', reloadTable).modal('hide'); successFlash("Template added successfully!"); list(); }).fail(x=>modalError((x.responseJSON&&x.responseJSON.message)||("Create failed ("+x.status+")"))); }
  });
  $(function(){ 
    // compat múltiplos modais
    $('.modal').on('hidden.bs.modal', function () {
      $(this).removeClass('fv-modal-stack');
      $('body').data('fv_open_modals', ($('body').data('fv_open_modals')||1) - 1);
    });
    $('.modal').on('shown.bs.modal', function () {
      if (typeof ($('body').data('fv_open_modals')) == 'undefined') $('body').data('fv_open_modals', 0);
      if ($(this).hasClass('fv-modal-stack')) return;
      $(this).addClass('fv-modal-stack');
      $('body').data('fv_open_modals', $('body').data('fv_open_modals') + 1);
      $(this).css('z-index', 1040 + (10 * $('body').data('fv_open_modals')));
      $('.modal-backdrop').not('.fv-modal-stack')
        .css('z-index', 1039 + (10 * $('body').data('fv_open_modals')))
        .addClass('fv-modal-stack');
    });
    $(document).on('hidden.bs.modal', '.modal', function () {
      $('.modal:visible').length && $(document.body).addClass('modal-open');
    });
    updateStats(); list(); initTooltips();});
 
})();
