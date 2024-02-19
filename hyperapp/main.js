import {h, text, app} from "https://cdn.skypack.dev/hyperapp"
import {users} from "./users.js"

app({
   init: {users},
   view: state => h("main", {}, [
      h("div", {id: "main"}, [
         h("h1", {}, text("Players")),
         playerList(state.users, addToQueue),
      ]),
      h("div", {id: "game"}),
      h("div", {id: "queue"}, [
         h("h1", {}, text("Queue")),
         h("div", {id: "queue-list"}),
      ]),
      h("div", {id: "recent"}, [
         h("div", {}, text("Recent")),
         h("div", {id: "recent-list"}),
      ]),
   ]),
   node: document.getElementById("app"),
})

function playerList(users, f) {
   return h("div", {id: "player-list"},
      users.map(u => addScore(u)).sort((a, b) => b.Score - a.Score).map((u, i) => playerRow(u, f, i+1))
   )
}

function playerRow(user, f, i) {
   return h("div", {class: "player", onclick: f(user.ID), key: i}, [
      h("span", {class: "index"}, text(i)),
      h("img", {class: "avatar", src: user.Avatar}),
      h("span", {class: "name"}, text([user.RealName, user.Name].filter(x => x !== "")[0])),
      h("span", {class: "skill"}, text(user.Score.toFixed(2) + " (" + user.Sigma.toFixed(2) + ")"))
   ])
}

function addToQueue(id) {
   return null
}

function addScore(user) {
   return {
      ...user,
      Score: user.Skill.Mu - 3 * Math.sqrt(user.Skill.SigSq),
      Sigma: Math.sqrt(user.Skill.SigSq),
   }
}
