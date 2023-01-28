const signoutBtn = document.getElementById("signout-btn")
const createRoomBtn = document.querySelector(".create-room-btn")
const memberName = document.getElementById("member-name")
nameOnNavbar()

signoutBtn.addEventListener("click", () => {
  fetch(
    "/api/user/auth"
  ).then(function (response) {
    return response.json();
  }).then(function (data) {
    if (data.data != undefined) {
      deleteAccount()
    }
  })
});

function nameOnNavbar() {
  fetch(
    "/api/user/auth"
  ).then(function (response) {
    return response.json();
  }).then(function (data) {
    memberName.textContent = "Hello, " + data.data.name
  })
}

createRoomBtn.addEventListener("click", () => {
  document.location.href = "/room/create"
})

async function deleteAccount() {
  let url = "/api/user/auth"
  let options = {
    method: "DELETE",
  }
  try {
    let response = await fetch(url, options);
    if (response.status === 200) {
      document.location.href = '/'
    }
  } catch (err) {
    console.log({ "error": err.message });
  }
}