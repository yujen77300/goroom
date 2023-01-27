const signoutBtn = document.getElementById("signout-btn")

signoutBtn.addEventListener("click", function () {
  fetch(
    "/api/user/auth"
  ).then(function (response) {
    return response.json();
  }).then(function (data) {
    if (data.data != undefined) {
      deleteAccount()
      console.log("要來刪除")
    }
  })
});

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