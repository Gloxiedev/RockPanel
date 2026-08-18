(function(){
"use strict";

var currentPage="overview";
var metrics={};
var refreshTimer=null;

var nav=[
  {id:"overview",label:"Overview",icon:"&#9673;"},
  {id:"servers",label:"Servers",icon:"&#9881;"},
  {id:"applications",label:"Applications",icon:"&#9776;"},
  {id:"docker",label:"Docker",icon:"&#9632;"},
  {id:"minecraft",label:"Minecraft",icon:"&#9618;"},
  {id:"websites",label:"Websites",icon:"&#9741;"},
  {id:"databases",label:"Databases",icon:"&#9635;"},
  {id:"files",label:"Files",icon:"&#9776;"},
  {id:"backups",label:"Backups",icon:"&#9850;"},
  {id:"schedules",label:"Schedules",icon:"&#9200;"},
  {id:"users",label:"Users",icon:"&#9787;"},
  {id:"tokens",label:"API Tokens",icon:"&#10003;"},
  {id:"logs",label:"Logs",icon:"&#9112;"},
  {id:"settings",label:"Settings",icon:"&#9881;"}
];

function api(method,url,data){
  return new Promise(function(resolve,reject){
    var xhr=new XMLHttpRequest();
    xhr.open(method,url,true);
    xhr.setRequestHeader("Content-Type","application/json");
    xhr.onload=function(){
      if(xhr.status>=200&&xhr.status<300){
        try{resolve(JSON.parse(xhr.responseText))}catch(e){resolve(xhr.responseText)}
      }else if(xhr.status===401){
        showLogin();
        reject("unauthorized");
      }else{
        reject(xhr.responseText);
      }
    };
    xhr.onerror=function(){reject("network error")};
    if(data)xhr.send(JSON.stringify(data));else xhr.send();
  });
}

function formatBytes(b){
  if(b===0)return"0 B";
  var u=["B","KB","MB","GB","TB"];
  var i=Math.floor(Math.log(b)/Math.log(1024));
  return(b/Math.pow(1024,i)).toFixed(1)+" "+u[i];
}

function formatSpeed(b){return formatBytes(b)+"/s"}

function render(){
  var app=document.getElementById("app");
  app.innerHTML=
    "<div class='sidebar' id='sidebar'>"+renderSidebar()+
    "<div class='main'><div class='topbar'>"+renderTopbar()+
    "</div><div class='content' id='content'></div></div>";

  document.querySelectorAll(".nav-link").forEach(function(el){
    el.addEventListener("click",function(){
      currentPage=this.dataset.page;
      render();
      loadPage();
    });
  });

  document.getElementById("logout-btn").addEventListener("click",function(){
    api("POST","/api/v1/auth/logout").then(function(){showLogin()});
  });

  loadPage();
}

function renderSidebar(){
  var h="<div class='sidebar-brand'>RockPanel</div><div class='sidebar-nav'>";
  nav.forEach(function(n){
    h+="<div class='nav-link"+(currentPage===n.id?" active":"")+"' data-page='"+n.id+"'>"+n.icon+" "+n.label+"</div>";
  });
  h+="</div>";
  return h;
}

function renderTopbar(){
  return "<h2>"+nav.find(function(n){return n.id===currentPage}).label+"</h2>"+
    "<div class='user-info'><span id='username'>admin</span><button class='btn btn-sm' id='logout-btn'>Logout</button></div>";
}

function loadPage(){
  api("GET","/api/v1/auth/me").then(function(u){
    if(u&&u.username){document.getElementById("username").textContent=u.username}
  }).catch(function(){});

  var c=document.getElementById("content");
  if(!c)return;

  switch(currentPage){
    case"overview":loadOverview(c);break;
    case"servers":loadServers(c);break;
    case"applications":loadApps(c);break;
    case"docker":loadDocker(c);break;
    case"minecraft":loadMinecraft(c);break;
    case"websites":loadWebsites(c);break;
    case"databases":loadDatabases(c);break;
    case"files":loadFiles(c);break;
    case"backups":loadBackups(c);break;
    case"schedules":loadSchedules(c);break;
    case"users":loadUsers(c);break;
    case"tokens":loadTokens(c);break;
    case"logs":loadLogs(c);break;
    case"settings":loadSettings(c);break;
    default:c.innerHTML="<div class='empty'>Page not found</div>";
  }
}

function loadOverview(c){
  api("GET","/api/v1/metrics").then(function(m){
    metrics=m;
    var cpuClass=m.cpu>90?"danger":m.cpu>70?"warn":"ok";
    var ramPct=m.ram.total>0?((m.ram.used/m.ram.total)*100):0;
    var ramClass=ramPct>90?"danger":ramPct>70?"warn":"ok";
    var diskPct=m.disk.total>0?((m.disk.used/m.disk.total)*100):0;
    var diskClass=diskPct>90?"danger":diskPct>70?"warn":"ok";

    c.innerHTML=
      "<div class='stats'>"+
        statCard("CPU",m.cpu.toFixed(1)+"%",cpuClass,m.cpu)+
        statCard("RAM",formatBytes(m.ram.used)+" / "+formatBytes(m.ram.total),ramClass,ramPct)+
        statCard("Disk",formatBytes(m.disk.used)+" / "+formatBytes(m.disk.total),diskClass,diskPct)+
        statCard("Network","&#8595;"+formatSpeed(m.network.rx_speed)+" &#8593;"+formatSpeed(m.network.tx_speed),"",0)+
        statCard("Load",m.load.load1.toFixed(2)+" / "+m.load.load5.toFixed(2)+" / "+m.load.load15.toFixed(2),"",0)+
        statCard("Uptime",formatUptime(m.uptime),"",0)+
      "</div>"+
      "<div class='card'><div class='card-head'><h3>System</h3></div><div class='card-body'>"+
        "<p>RockPanel v0.1.0 running</p>"+
      "</div></div>";

    if(refreshTimer)clearInterval(refreshTimer);
    refreshTimer=setInterval(function(){loadOverview(c)},5000);
  }).catch(function(err){
    c.innerHTML="<div class='empty'>Failed to load metrics: "+err+"</div>";
  });
}

function statCard(label,value,cls,pct){
  var bar=pct>0?"<div class='progress-bar'><div class='progress-fill' style='width:"+Math.min(pct,100)+"%;background:var(--"+(cls==="danger"?"red":cls==="warn"?"yellow":"green")+")'></div></div>":"";
  return "<div class='stat'><div class='stat-label'>"+label+"</div><div class='stat-val "+cls+"'>"+value+"</div>"+bar+"</div>";
}

function formatUptime(s){
  var d=Math.floor(s/86400);
  var h=Math.floor((s%86400)/3600);
  var m=Math.floor((s%3600)/60);
  if(d>0)return d+"d "+h+"h";
  if(h>0)return h+"h "+m+"m";
  return m+"m";
}

function loadServers(c){
  api("GET","/api/v1/servers").then(function(servers){
    var h="<div class='card'><div class='card-head'><h3>Servers</h3><button class='btn btn-primary' onclick='RP.createServer()'>+ New Server</button></div><div class='card-body'><table><tr><th>ID</th><th>Name</th><th>Status</th><th>PID</th><th>Actions</th></tr>";
    if(!servers||servers.length===0){
      h+="<tr><td colspan='5' class='empty'>No servers configured</td></tr>";
    }else{
      servers.forEach(function(s){
        h+="<tr><td>"+s.id+"</td><td>"+esc(s.name)+"</td><td><span class='badge badge-"+(s.status==="running"?"on":"off")+"'>"+s.status+"</span></td><td>"+s.pid+"</td><td class='btn-group'>"+
          (s.status==="running"?
            "<button class='btn btn-sm' onclick='RP.stopServer("+s.id+")'>Stop</button>":
            "<button class='btn btn-sm btn-primary' onclick='RP.startServer("+s.id+")'>Start</button>")+
          "<button class='btn btn-sm btn-danger' onclick='RP.deleteServer("+s.id+")'>Delete</button></td></tr>";
      });
    }
    h+="</table></div></div>";
    c.innerHTML=h;
  });
}

function loadApps(c){
  api("GET","/api/v1/applications").then(function(apps){
    var h="<div class='card'><div class='card-head'><h3>Applications</h3><button class='btn btn-primary' onclick='RP.createApp()'>+ New App</button></div><div class='card-body'><table><tr><th>ID</th><th>Name</th><th>Status</th><th>Port</th><th>Actions</th></tr>";
    if(!apps||apps.length===0){
      h+="<tr><td colspan='5' class='empty'>No applications configured</td></tr>";
    }else{
      apps.forEach(function(a){
        h+="<tr><td>"+a.id+"</td><td>"+esc(a.name)+"</td><td><span class='badge badge-"+(a.status==="running"?"on":"off")+"'>"+a.status+"</span></td><td>"+a.port+"</td><td class='btn-group'>"+
          (a.status==="running"?
            "<button class='btn btn-sm' onclick='RP.stopApp("+a.id+")'>Stop</button>":
            "<button class='btn btn-sm btn-primary' onclick='RP.startApp("+a.id+")'>Start</button>")+
          "<button class='btn btn-sm btn-danger' onclick='RP.deleteApp("+a.id+")'>Delete</button></td></tr>";
      });
    }
    h+="</table></div></div>";
    c.innerHTML=h;
  });
}

function loadMinecraft(c){
  api("GET","/api/v1/minecraft").then(function(servers){
    var h="<div class='card'><div class='card-head'><h3>Minecraft Servers</h3><button class='btn btn-primary' onclick='RP.createMC()'>+ New Server</button></div><div class='card-body'><table><tr><th>ID</th><th>Name</th><th>Type</th><th>Version</th><th>Status</th><th>Actions</th></tr>";
    if(!servers||servers.length===0){
      h+="<tr><td colspan='6' class='empty'>No Minecraft servers</td></tr>";
    }else{
      servers.forEach(function(s){
        h+="<tr><td>"+s.id+"</td><td>"+esc(s.name)+"</td><td>"+esc(s.server_type)+"</td><td>"+esc(s.version)+"</td><td><span class='badge badge-"+(s.status==="running"?"on":"off")+"'>"+s.status+"</span></td><td class='btn-group'>"+
          (s.status==="running"?
            "<button class='btn btn-sm' onclick='RP.stopMC("+s.id+")'>Stop</button>":
            "<button class='btn btn-sm btn-primary' onclick='RP.startMC("+s.id+")'>Start</button>")+
          "<button class='btn btn-sm btn-danger' onclick='RP.deleteMC("+s.id+")'>Delete</button></td></tr>";
      });
    }
    h+="</table></div></div>";
    c.innerHTML=h;
  });
}

function loadDocker(c){
  api("GET","/api/v1/docker/containers").then(function(containers){
    var h="<div class='card'><div class='card-head'><h3>Docker Containers</h3></div><div class='card-body'><table><tr><th>ID</th><th>Name</th><th>Image</th><th>Status</th><th>Actions</th></tr>";
    if(!containers||containers.length===0){
      h+="<tr><td colspan='5' class='empty'>No containers running (Docker not available or empty)</td></tr>";
    }else{
      containers.forEach(function(ct){
        h+="<tr><td>"+esc(ct.id.substr(0,12))+"</td><td>"+esc(ct.name)+"</td><td>"+esc(ct.image)+"</td><td><span class='badge badge-"+(ct.state==="running"?"on":"off")+"'>"+ct.status+"</span></td><td class='btn-group'>"+
          "<button class='btn btn-sm'>Stop</button><button class='btn btn-sm'>Restart</button></td></tr>";
      });
    }
    h+="</table></div></div>";
    c.innerHTML=h;
  });
}

function loadWebsites(c){
  api("GET","/api/v1/websites").then(function(sites){
    var h="<div class='card'><div class='card-head'><h3>Websites</h3><button class='btn btn-primary' onclick='RP.createWebsite()'>+ New Website</button></div><div class='card-body'><table><tr><th>Domain</th><th>Target</th><th>SSL</th><th>Actions</th></tr>";
    if(!sites||sites.length===0){
      h+="<tr><td colspan='4' class='empty'>No websites configured</td></tr>";
    }else{
      sites.forEach(function(s){
        h+="<tr><td>"+esc(s.domain)+"</td><td>"+esc(s.target)+"</td><td><span class='badge badge-"+(s.ssl_enabled?"on":"off")+"'>"+(s.ssl_enabled?"Yes":"No")+"</span></td><td><button class='btn btn-sm btn-danger' onclick='RP.deleteWebsite("+s.id+")'>Delete</button></td></tr>";
      });
    }
    h+="</table></div></div>";
    c.innerHTML=h;
  });
}

function loadDatabases(c){
  api("GET","/api/v1/databases").then(function(dbs){
    var h="<div class='card'><div class='card-head'><h3>Databases</h3><button class='btn btn-primary' onclick='RP.createDB()'>+ New Database</button></div><div class='card-body'><table><tr><th>Name</th><th>Type</th><th>Host</th><th>Port</th><th>Actions</th></tr>";
    if(!dbs||dbs.length===0){
      h+="<tr><td colspan='5' class='empty'>No databases configured</td></tr>";
    }else{
      dbs.forEach(function(d){
        h+="<tr><td>"+esc(d.name)+"</td><td>"+esc(d.type)+"</td><td>"+esc(d.host)+"</td><td>"+d.port+"</td><td><button class='btn btn-sm btn-danger' onclick='RP.deleteDB("+d.id+")'>Delete</button></td></tr>";
      });
    }
    h+="</table></div></div>";
    c.innerHTML=h;
  });
}

function loadFiles(c){
  api("GET","/api/v1/files?path=").then(function(entries){
    var h="<div class='card'><div class='card-head'><h3>File Manager</h3></div><div class='card-body'><table><tr><th>Name</th><th>Size</th><th>Modified</th><th>Actions</th></tr>";
    if(!entries||entries.length===0){
      h+="<tr><td colspan='4' class='empty'>No files</td></tr>";
    }else{
      entries.forEach(function(f){
        var date=f.mod_time?new Date(f.mod_time*1000).toLocaleString():"";
        h+="<tr><td>"+(f.is_dir?"&#128193; ":"")+esc(f.name)+"</td><td>"+(f.is_dir?"-":formatBytes(f.size))+"</td><td>"+date+"</td><td class='btn-group'>"+
          "<button class='btn btn-sm btn-danger' onclick='RP.deleteFile(\""+esc(f.path)+"\")'>Delete</button></td></tr>";
      });
    }
    h+="</table></div></div>";
    c.innerHTML=h;
  });
}

function loadBackups(c){
  api("GET","/api/v1/backups").then(function(backups){
    var h="<div class='card'><div class='card-head'><h3>Backups</h3><button class='btn btn-primary' onclick='RP.createBackup()'>+ New Backup</button></div><div class='card-body'><table><tr><th>Name</th><th>Type</th><th>Status</th><th>Created</th><th>Actions</th></tr>";
    if(!backups||backups.length===0){
      h+="<tr><td colspan='5' class='empty'>No backups</td></tr>";
    }else{
      backups.forEach(function(b){
        var date=b.created_at?new Date(b.created_at*1000).toLocaleString():"";
        h+="<tr><td>"+esc(b.name)+"</td><td>"+esc(b.type)+"</td><td><span class='badge badge-on'>"+b.status+"</span></td><td>"+date+"</td><td><button class='btn btn-sm btn-danger' onclick='RP.deleteBackup("+b.id+")'>Delete</button></td></tr>";
      });
    }
    h+="</table></div></div>";
    c.innerHTML=h;
  });
}

function loadSchedules(c){
  api("GET","/api/v1/schedules").then(function(schedules){
    var h="<div class='card'><div class='card-head'><h3>Schedules</h3></div><div class='card-body'><table><tr><th>Name</th><th>Type</th><th>Cron</th><th>Enabled</th></tr>";
    if(!schedules||schedules.length===0){
      h+="<tr><td colspan='4' class='empty'>No schedules</td></tr>";
    }else{
      schedules.forEach(function(s){
        h+="<tr><td>"+esc(s.name)+"</td><td>"+esc(s.type)+"</td><td>"+esc(s.cron_expr)+"</td><td><span class='badge badge-"+(s.enabled?"on":"off")+"'>"+(s.enabled?"Yes":"No")+"</span></td></tr>";
      });
    }
    h+="</table></div></div>";
    c.innerHTML=h;
  });
}

function loadUsers(c){
  api("GET","/api/v1/users").then(function(users){
    var h="<div class='card'><div class='card-head'><h3>Users</h3><button class='btn btn-primary' onclick='RP.createUser()'>+ New User</button></div><div class='card-body'><table><tr><th>ID</th><th>Username</th><th>Role</th><th>Actions</th></tr>";
    if(!users||users.length===0){
      h+="<tr><td colspan='4' class='empty'>No users</td></tr>";
    }else{
      users.forEach(function(u){
        h+="<tr><td>"+u.id+"</td><td>"+esc(u.username)+"</td><td>"+esc(u.role)+"</td><td><button class='btn btn-sm btn-danger' onclick='RP.deleteUser("+u.id+")'>Delete</button></td></tr>";
      });
    }
    h+="</table></div></div>";
    c.innerHTML=h;
  });
}

function loadTokens(c){
  api("GET","/api/v1/tokens").then(function(tokens){
    var h="<div class='card'><div class='card-head'><h3>API Tokens</h3><button class='btn btn-primary' onclick='RP.createToken()'>+ New Token</button></div><div class='card-body'><table><tr><th>Name</th><th>Prefix</th><th>Created</th><th>Last Used</th><th>Actions</th></tr>";
    if(!tokens||tokens.length===0){
      h+="<tr><td colspan='5' class='empty'>No tokens</td></tr>";
    }else{
      tokens.forEach(function(t){
        var created=t.created_at?new Date(t.created_at*1000).toLocaleString():"";
        var lastUsed=t.last_used?new Date(t.last_used*1000).toLocaleString():"Never";
        h+="<tr><td>"+esc(t.name)+"</td><td>"+esc(t.prefix)+"...</td><td>"+created+"</td><td>"+lastUsed+"</td><td><button class='btn btn-sm btn-danger' onclick='RP.revokeToken("+t.id+")'>Revoke</button></td></tr>";
      });
    }
    h+="</table></div></div>";
    c.innerHTML=h;
  });
}

function loadLogs(c){
  c.innerHTML="<div class='card'><div class='card-head'><h3>Panel Logs</h3></div><div class='card-body'><div class='console' id='log-console'>No log entries yet.</div></div></div>";
}

function loadSettings(c){
  c.innerHTML="<div class='card'><div class='card-head'><h3>Settings</h3></div><div class='card-body'>"+
    "<p style='color:var(--fg2)'>Configuration is managed via the config file.</p>"+
    "<p style='margin-top:8px'>Edit /etc/rockpanel/config.yaml to change settings.</p>"+
    "</div></div>";
}

function esc(s){if(!s)return"";var d=document.createElement("div");d.textContent=s;return d.innerHTML}

function showModal(title,bodyHtml,footerHtml){
  var m=document.createElement("div");
  m.className="modal-bg";
  m.id="modal";
  m.innerHTML="<div class='modal'><div class='modal-head'><h3>"+title+"</h3><button class='modal-x' onclick='RP.closeModal()'>&times;</button></div><div class='modal-body'>"+bodyHtml+"</div>"+(footerHtml?"<div class='modal-foot'>"+footerHtml+"</div>":"")+"</div>";
  document.body.appendChild(m);
  setTimeout(function(){m.classList.add("open")},10);
}

function closeModal(){
  var m=document.getElementById("modal");
  if(m){m.classList.remove("open");setTimeout(function(){m.remove()},200)}
}

function showLogin(){
  if(refreshTimer)clearInterval(refreshTimer);
  var app=document.getElementById("app");
  app.innerHTML="<div style='display:flex;align-items:center;justify-content:center;min-height:100vh'>"+
    "<div class='modal' style='position:static;opacity:1;visibility:visible;transform:none'>"+
    "<div class='modal-head'><h3>Login to RockPanel</h3></div>"+
    "<div class='modal-body'>"+
    "<div class='form-group'><label class='form-label'>Username</label><input class='form-input' id='login-user' value='admin'></div>"+
    "<div class='form-group'><label class='form-label'>Password</label><input class='form-input' id='login-pass' type='password'></div>"+
    "</div><div class='modal-foot'><button class='btn btn-primary' id='login-btn'>Login</button></div></div></div>";

  document.getElementById("login-btn").addEventListener("click",function(){
    var u=document.getElementById("login-user").value;
    var p=document.getElementById("login-pass").value;
    api("POST","/api/v1/auth/login",{username:u,password:p}).then(function(){
      render();
    }).catch(function(){
      alert("Invalid credentials");
    });
  });
  document.getElementById("login-pass").addEventListener("keydown",function(e){
    if(e.key==="Enter")document.getElementById("login-btn").click();
  });
}

window.RP={
  closeModal:closeModal,
  createServer:function(){
    showModal("New Server",
      "<div class='form-group'><label class='form-label'>Name</label><input class='form-input' id='srv-name'></div>"+
      "<div class='form-group'><label class='form-label'>Command</label><input class='form-input' id='srv-cmd' placeholder='/usr/bin/node server.js'></div>"+
      "<div class='form-row'><div class='form-group'><label class='form-label'>Work Directory</label><input class='form-input' id='srv-dir'></div>"+
      "<div class='form-group'><label class='form-label'>Restart Policy</label><select class='form-select' id='srv-restart'><option value='on-failure'>On Failure</option><option value='always'>Always</option><option value='never'>Never</option></select></div></div>"+
      "<div class='form-group'><label class='form-label'>Environment Variables</label><textarea class='form-textarea' id='srv-env' placeholder='KEY=VALUE'></textarea></div>",
      "<button class='btn' onclick='RP.closeModal()'>Cancel</button><button class='btn btn-primary' onclick='RP.doCreateServer()'>Create</button>"
    );
  },
  doCreateServer:function(){
    api("POST","/api/v1/servers",{name:gi("srv-name"),command:gi("srv-cmd"),work_dir:gi("srv-dir"),restart_policy:gi("srv-restart"),env_vars:gi("srv-env")}).then(function(){closeModal();loadPage()});
  },
  startServer:function(id){api("POST","/api/v1/servers/"+id+"/start").then(function(){loadPage()})},
  stopServer:function(id){api("POST","/api/v1/servers/"+id+"/stop").then(function(){loadPage()})},
  deleteServer:function(id){if(confirm("Delete server?"))api("DELETE","/api/v1/servers/"+id).then(function(){loadPage()})},

  createApp:function(){
    showModal("New Application",
      "<div class='form-group'><label class='form-label'>Name</label><input class='form-input' id='app-name'></div>"+
      "<div class='form-group'><label class='form-label'>Git Repository</label><input class='form-input' id='app-repo' placeholder='https://github.com/...'></div>"+
      "<div class='form-row'><div class='form-group'><label class='form-label'>Branch</label><input class='form-input' id='app-branch' value='main'></div>"+
      "<div class='form-group'><label class='form-label'>Port</label><input class='form-input' id='app-port' type='number'></div></div>"+
      "<div class='form-group'><label class='form-label'>Install Command</label><input class='form-input' id='app-install' placeholder='npm install'></div>"+
      "<div class='form-group'><label class='form-label'>Build Command</label><input class='form-input' id='app-build' placeholder='npm run build'></div>"+
      "<div class='form-group'><label class='form-label'>Start Command</label><input class='form-input' id='app-start' placeholder='npm start'></div>"+
      "<div class='form-group'><label class='form-label'>Working Directory</label><input class='form-input' id='app-dir'></div>",
      "<button class='btn' onclick='RP.closeModal()'>Cancel</button><button class='btn btn-primary' onclick='RP.doCreateApp()'>Create</button>"
    );
  },
  doCreateApp:function(){
    var port=parseInt(gi("app-port"))||0;
    api("POST","/api/v1/applications",{name:gi("app-name"),git_repo:gi("app-repo"),branch:gi("app-branch"),work_dir:gi("app-dir"),install_cmd:gi("app-install"),build_cmd:gi("app-build"),start_cmd:gi("app-start"),port:port}).then(function(){closeModal();loadPage()});
  },
  startApp:function(id){api("POST","/api/v1/applications/"+id+"/start").then(function(){loadPage()})},
  stopApp:function(id){api("POST","/api/v1/applications/"+id+"/stop").then(function(){loadPage()})},
  deleteApp:function(id){if(confirm("Delete application?"))api("DELETE","/api/v1/applications/"+id).then(function(){loadPage()})},

  createMC:function(){
    showModal("New Minecraft Server",
      "<div class='form-group'><label class='form-label'>Name</label><input class='form-input' id='mc-name'></div>"+
      "<div class='form-row'><div class='form-group'><label class='form-label'>Server Type</label><select class='form-select' id='mc-type'><option>vanilla</option><option>paper</option><option>purpur</option><option>fabric</option><option>forge</option></select></div>"+
      "<div class='form-group'><label class='form-label'>Version</label><input class='form-input' id='mc-version' value='1.20.4'></div></div>"+
      "<div class='form-row'><div class='form-group'><label class='form-label'>Memory (MB)</label><input class='form-input' id='mc-memory' type='number' value='1024'></div>"+
      "<div class='form-group'><label class='form-label'>Port</label><input class='form-input' id='mc-port' type='number' value='25565'></div></div>"+
      "<div class='form-group'><label class='form-label'>Java Version</label><select class='form-select' id='mc-java'><option value='java8'>Java 8</option><option value='java17' selected>Java 17</option><option value='java21'>Java 21</option><option value='java25'>Java 25</option></select></div>",
      "<button class='btn' onclick='RP.closeModal()'>Cancel</button><button class='btn btn-primary' onclick='RP.doCreateMC()'>Create</button>"
    );
  },
  doCreateMC:function(){
    api("POST","/api/v1/minecraft",{name:gi("mc-name"),server_type:gi("mc-type"),version:gi("mc-version"),memory:parseInt(gi("mc-memory"))||1024,port:parseInt(gi("mc-port"))||25565,java_version:gi("mc-java")}).then(function(){closeModal();loadPage()});
  },
  startMC:function(id){api("POST","/api/v1/minecraft/"+id+"/start").then(function(){loadPage()})},
  stopMC:function(id){api("POST","/api/v1/minecraft/"+id+"/stop").then(function(){loadPage()})},
  deleteMC:function(id){if(confirm("Delete Minecraft server?"))api("DELETE","/api/v1/minecraft/"+id).then(function(){loadPage()})},

  createWebsite:function(){
    showModal("New Website",
      "<div class='form-group'><label class='form-label'>Domain</label><input class='form-input' id='web-domain' placeholder='example.com'></div>"+
      "<div class='form-group'><label class='form-label'>Target</label><input class='form-input' id='web-target' placeholder='localhost:3000'></div>"+
      "<div class='form-group'><label class='form-label'>Enable SSL</label><select class='form-select' id='web-ssl'><option value='false'>No</option><option value='true'>Yes</option></select></div>",
      "<button class='btn' onclick='RP.closeModal()'>Cancel</button><button class='btn btn-primary' onclick='RP.doCreateWebsite()'>Create</button>"
    );
  },
  doCreateWebsite:function(){
    api("POST","/api/v1/websites",{domain:gi("web-domain"),target:gi("web-target"),ssl_enabled:gi("web-ssl")==="true"}).then(function(){closeModal();loadPage()});
  },
  deleteWebsite:function(id){if(confirm("Delete website?"))api("DELETE","/api/v1/websites/"+id).then(function(){loadPage()})},

  createDB:function(){
    showModal("New Database",
      "<div class='form-group'><label class='form-label'>Name</label><input class='form-input' id='db-name'></div>"+
      "<div class='form-row'><div class='form-group'><label class='form-label'>Type</label><select class='form-select' id='db-type'><option>sqlite</option><option>postgresql</option><option>mysql</option></select></div>"+
      "<div class='form-group'><label class='form-label'>Port</label><input class='form-input' id='db-port' type='number' value='5432'></div></div>"+
      "<div class='form-row'><div class='form-group'><label class='form-label'>Username</label><input class='form-input' id='db-user'></div>"+
      "<div class='form-group'><label class='form-label'>Password</label><input class='form-input' id='db-pass' type='password'></div></div>",
      "<button class='btn' onclick='RP.closeModal()'>Cancel</button><button class='btn btn-primary' onclick='RP.doCreateDB()'>Create</button>"
    );
  },
  doCreateDB:function(){
    api("POST","/api/v1/databases",{name:gi("db-name"),type:gi("db-host"),host:"localhost",port:parseInt(gi("db-port"))||0,username:gi("db-user"),password:gi("db-pass")}).then(function(){closeModal();loadPage()});
  },
  deleteDB:function(id){if(confirm("Delete database?"))api("DELETE","/api/v1/databases/"+id).then(function(){loadPage()})},

  createBackup:function(){
    showModal("New Backup",
      "<div class='form-group'><label class='form-label'>Name</label><input class='form-input' id='bk-name' placeholder='backup-2024-01-01'></div>"+
      "<div class='form-group'><label class='form-label'>Type</label><select class='form-select' id='bk-type'><option>manual</option><option>server</option><option>application</option><option>database</option></select></div>",
      "<button class='btn' onclick='RP.closeModal()'>Cancel</button><button class='btn btn-primary' onclick='RP.doCreateBackup()'>Create</button>"
    );
  },
  doCreateBackup:function(){
    api("POST","/api/v1/backups",{name:gi("bk-name"),type:gi("bk-type")}).then(function(){closeModal();loadPage()});
  },
  deleteBackup:function(id){if(confirm("Delete backup?"))api("DELETE","/api/v1/backups/"+id).then(function(){loadPage()})},

  createUser:function(){
    showModal("New User",
      "<div class='form-group'><label class='form-label'>Username</label><input class='form-input' id='usr-name'></div>"+
      "<div class='form-group'><label class='form-label'>Password</label><input class='form-input' id='usr-pass' type='password'></div>"+
      "<div class='form-group'><label class='form-label'>Role</label><select class='form-select' id='usr-role'><option value='user'>User</option><option value='admin'>Admin</option></select></div>",
      "<button class='btn' onclick='RP.closeModal()'>Cancel</button><button class='btn btn-primary' onclick='RP.doCreateUser()'>Create</button>"
    );
  },
  doCreateUser:function(){
    api("POST","/api/v1/users",{username:gi("usr-name"),password:gi("usr-pass"),role:gi("usr-role")}).then(function(){closeModal();loadPage()});
  },
  deleteUser:function(id){if(confirm("Delete user?"))api("DELETE","/api/v1/users/"+id).then(function(){loadPage()})},

  createToken:function(){
    showModal("New API Token",
      "<div class='form-group'><label class='form-label'>Name</label><input class='form-input' id='tkn-name'></div>",
      "<button class='btn' onclick='RP.closeModal()'>Cancel</button><button class='btn btn-primary' onclick='RP.doCreateToken()'>Create</button>"
    );
  },
  doCreateToken:function(){
    api("POST","/api/v1/tokens",{name:gi("tkn-name")}).then(function(resp){
      closeModal();
      showModal("Token Created",
        "<div class='form-group'><label class='form-label'>Token</label><input class='form-input' value='"+esc(resp.token)+"' readonly></div>"+
        "<p style='color:var(--yellow);font-size:13px;margin-top:8px'>Save this token — it will not be shown again.</p>",
        "<button class='btn btn-primary' onclick='RP.closeModal()'>Close</button>"
      );
      loadPage();
    });
  },
  revokeToken:function(id){if(confirm("Revoke token?"))api("DELETE","/api/v1/tokens/"+id).then(function(){loadPage()})},

  deleteFile:function(path){if(confirm("Delete file?"))api("DELETE","/api/v1/files/"+encodeURIComponent(path)).then(function(){loadPage()})}
};

function gi(id){return document.getElementById(id).value}

api("GET","/api/v1/auth/me").then(function(u){
  if(u&&u.id){render()}else{showLogin()}
}).catch(function(){showLogin()});

})();
