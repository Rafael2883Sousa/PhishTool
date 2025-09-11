/* global $, user, csrf_token */
(function () {
  function escapeHtml(s){
    return s.replace(/[&<>"']/g, m => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[m]));
  }

  function authHeaders() {
    const h = {};
    if (window.user && user.api_key) h["Authorization"] = "Bearer " + user.api_key;
    if (window.csrf_token) h["X-CSRF-Token"] = csrf_token; // se houver middleware CSRF
    return h;
  }

  function api(url, opts) {
    const o = Object.assign({ method: "GET" }, opts || {});
    o.url = url;
    o.headers = Object.assign({}, authHeaders(), o.headers || {});
    return $.ajax(o);
  }

  function row(p){
    return `<tr data-id="${p.id}">
      <td>${p.id}</td>
      <td>${escapeHtml(p.name)}</td>
      <td>${escapeHtml(p.from_number)}</td>
      <td>${p.rate_limit_per_min}</td>
      <td>
        <button class="btn btn-sm btn-default edit">Edit</button>
        <button class="btn btn-sm btn-danger delete">Delete</button>
      </td></tr>`;
  }

  function list(){
    api("/api/sms/profiles/", { dataType: "json" })
      .done(items => {
        const $tb = $("#tbl-profiles tbody").empty();
        items.forEach(p => $tb.append(row(p)));
      })
      .fail(xhr => alert("List failed: " + xhr.status + " " + (xhr.responseText || "")));
  }

  $("#form-profile").on("submit", function(e){
    e.preventDefault();
    const payload = {
      name: this.name.value.trim(),
      account_sid: this.account_sid.value.trim(),
      auth_token: this.auth_token.value.trim(),
      from_number: this.from_number.value.trim(),
      rate_limit_per_min: parseInt(this.rate_limit_per_min.value || "60", 10)
    };
    api("/api/sms/profiles/", {
      method: "POST",
      contentType: "application/json",
      data: JSON.stringify(payload)
    })
    .done(()=>{ this.reset(); list(); })
    .fail(xhr=>alert("Create failed: " + xhr.status + " " + (xhr.responseText || "")));
  });

  $("#tbl-profiles").on("click","button.delete", function(){
    const id = $(this).closest("tr").data("id");
    if(!confirm("Delete profile #"+id+"?")) return;
    api("/api/sms/profiles/" + id, { method: "DELETE" })
      .done(list)
      .fail(xhr=>alert("Delete failed: " + xhr.status + " " + (xhr.responseText || "")));
  });

  $("#tbl-profiles").on("click","button.edit", function(){
    const $tr = $(this).closest("tr");
    const id = $tr.data("id");
    const name = prompt("Name:", $tr.children().eq(1).text());
    if(name==null) return;
    const rate = prompt("Rate/min:", $tr.children().eq(3).text());
    if(rate==null) return;
    api("/api/sms/profiles/" + id, {
      method: "PUT",
      contentType: "application/json",
      data: JSON.stringify({ name, rate_limit_per_min: parseInt(rate,10)||60 })
    })
    .done(list)
    .fail(xhr=>alert("Update failed: " + xhr.status + " " + (xhr.responseText || "")));
  });

  $(list);
})();
