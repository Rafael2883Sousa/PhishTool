/* global $, csrf_token */
var smsProfiles = [];

(function () {
  // ---- helpers ----
  function esc(s){return String(s||"").replace(/[&<>"']/g,m=>({"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;"}[m]));}
  function isE164(s){return /^\+[1-9]\d{6,14}$/.test(s);}
  function auth(){const h={},u=window.user||{}; if(u.api_key) h.Authorization="Bearer "+u.api_key; if(window.csrf_token) h["X-CSRF-Token"]=csrf_token; return h;}
  function api(url,opt){opt=opt||{}; opt.url=url; opt.headers=Object.assign({},auth(),opt.headers||{}); return $.ajax(opt);}

  // ---- CRUD ----
  function save(idx){
    var p = {
      name: $("#name").val().trim(),
      account_sid: $("#account_sid").val().trim(),
      from_number: $("#from_number").val().trim(),
      rate_limit_per_min: parseInt($("#rate_limit_per_min").val()||"60",10)
    };
    var tok = $("#auth_token").val().trim();
    if (tok) p.auth_token = tok;

    if(!p.name || !p.account_sid || !isE164(p.from_number) || !(p.rate_limit_per_min>0)){
      modalError("Invalid fields. Check Account SID, From (E.164) and Rate.");
      return;
    }

    if (idx !== -1){
      p.id = smsProfiles[idx].id;
      api("/api/sms/profiles/"+p.id,{method:"PUT",contentType:"application/json",data:JSON.stringify(p)})
        .done(function(){ successFlash("Profile edited successfully!"); load(); dismiss(); })
        .fail(function(x){ modalError((x.responseJSON&&x.responseJSON.message)||"Update failed"); });
    } else {
      api("/api/sms/profiles/",{method:"POST",contentType:"application/json",data:JSON.stringify(p)})
        .done(function(){ successFlash("Profile added successfully!"); load(); dismiss(); })
        .fail(function(x){ modalError((x.responseJSON&&x.responseJSON.message)||"Create failed"); });
    }
  }

  function dismiss(){
    $("#modal\\.flashes").empty();
    $("#name").val("");
    $("#account_sid").val("");
    $("#auth_token").val("");
    $("#from_number").val("");
    $("#rate_limit_per_min").val("60");
    $("#modal").modal("hide");
  }

  window.edit = function(idx){
    $("#modalSubmit").off("click").on("click", function(){ save(idx); });
    if (idx !== -1){
      $("#profileModalLabel").text("Edit SMS Profile");
      var p = smsProfiles[idx];
      $("#name").val(p.name);
      $("#account_sid").val(p.account_sid);
      $("#auth_token").val("");
      $("#from_number").val(p.from_number);
      $("#rate_limit_per_min").val(p.rate_limit_per_min);
    } else {
      $("#profileModalLabel").text("New SMS Profile");
      dismiss(); // limpa e fecha; reabrimos já abaixo
      $("#modal").modal("show");
      $("#profileModalLabel").text("New SMS Profile");
      return;
    }
    $("#modal").modal("show");
  };

  window.deleteProfile = function(idx){
    Swal.fire({
      title: "Are you sure?",
      text: "This will delete the SMS profile. This can't be undone!",
      type: "warning",
      animation: false,
      showCancelButton: true,
      confirmButtonText: "Delete " + esc(smsProfiles[idx].name),
      confirmButtonColor: "#428bca",
      reverseButtons: true,
      allowOutsideClick: false,
      preConfirm: function () {
        return new Promise(function (resolve, reject) {
          api("/api/sms/profiles/"+smsProfiles[idx].id,{method:"DELETE"})
            .done(function(){ resolve(); })
            .fail(function(x){ reject((x.responseJSON&&x.responseJSON.message)||"Delete failed"); });
        });
      }
    }).then(function (result) {
      if (result.value){
        Swal.fire('SMS Profile Deleted!','This SMS profile has been deleted!','success');
      }
      $('button:contains("OK")').on('click', function () { location.reload(); });
    });
  };

  // ---- load table ----
  function load(){
    $("#profileTable").hide();
    $("#emptyMessage").hide();
    $("#loading").show();

    api("/api/sms/profiles/",{method:"GET",dataType:"json"})
      .done(function(list){
        smsProfiles = list || [];
        $("#loading").hide();
        if (smsProfiles.length > 0){
          $("#profileTable").show();
          var dt = $("#profileTable").DataTable({
            destroy:true,
            columnDefs:[{orderable:false, targets:"no-sort"}]
          });
          dt.clear();
          var rows=[];
          $.each(smsProfiles, function(i,p){
            rows.push([
              esc(p.name),
              esc(p.from_number),
              p.rate_limit_per_min,
              "<div class='pull-right'>\
                 <span data-toggle='modal' data-backdrop='static' data-target='#modal'>\
                   <button class='btn btn-primary' data-toggle='tooltip' data-placement='left' title='Edit Profile' onclick='edit("+i+")'>\
                     <i class='fa fa-pencil'></i>\
                   </button></span>\
                 <button class='btn btn-danger' data-toggle='tooltip' data-placement='left' title='Delete Profile' onclick='deleteProfile("+i+")'>\
                   <i class='fa fa-trash-o'></i>\
                 </button></div>"
            ]);
          });
          dt.rows.add(rows).draw();
          $('[data-toggle=\"tooltip\"]').tooltip();
        } else {
          $("#emptyMessage").show();
        }
      })
      .fail(function(){ $("#loading").hide(); errorFlash("Error fetching SMS profiles"); });
  }

  // ---- utilities like in Gophish ----
  window.modalError = function(msg){
    $("#modal\\.flashes").empty().append(
      '<div style="text-align:center" class="alert alert-danger"><i class="fa fa-exclamation-circle"></i> '+esc(msg)+'</div>');
  };
  window.successFlash = function(msg){
    $("#flashes").empty().append(
      '<div style="text-align:center" class="alert alert-success"><i class="fa fa-check-circle"></i> '+esc(msg)+'</div>');
    setTimeout(function(){ $("#flashes").empty(); }, 4000);
  };
  window.errorFlash = function(msg){
    $("#flashes").empty().append(
      '<div style="text-align:center" class="alert alert-danger"><i class="fa fa-exclamation-circle"></i> '+esc(msg)+'</div>');
  };

  // ---- boot ----
  $(document).ready(function(){
    // compat múltiplos modais (igual ao sending_profiles)
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
      $('.modal-backdrop').not('.fv-modal-stack').css('z-index', 1039 + (10 * $('body').data('fv_open_modals'))).addClass('fv-modal-stack');
    });
    $(document).on('hidden.bs.modal', '.modal', function () {
      $('.modal:visible').length && $(document.body).addClass('modal-open');
    });

    $("#modal").on("hidden.bs.modal", function () { dismiss(); });
    $("#modalSubmit").on("click", function(){ save(-1); }); // default action for "New Profile"

    load();
  });
})();
