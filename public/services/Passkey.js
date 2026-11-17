export const Passkeys = {
  register: async (username) => {
    try {
      const response = await fetch("/api/passkey/registration-begin", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: app.Store.jwt ? `Bearer ${app.Store.jwt}` : null,
        },
        body: JSON.stringify({ username: username }),
      });

      // Server errors are plain text (http.Error), not JSON.
      if (!response.ok) {
        const err = await response.text();
        app.showError("Failed to get registration options from server. " + err);
        return;
      }

      const options = await response.json();

      const attestationResponse = await SimpleWebAuthnBrowser.startRegistration(
        { optionsJSON: options.publicKey },
      );

      const verificationResponse = await fetch(
        "/api/passkey/registration-end",
        {
          method: "POST",
          credentials: "same-origin",
          headers: {
            "Content-Type": "application/json",
            Authorization: app.Store.jwt ? `Bearer ${app.Store.jwt}` : null,
          },
          body: JSON.stringify(attestationResponse),
        },
      );

      if (verificationResponse.ok) {
        app.showError(
          "Your passkey was saved. You can use it next time to login",
        );
      } else {
        app.showError(await verificationResponse.text(), false);
      }
    } catch (e) {
      app.showError("Error: " + e.message, false);
    }
  },
  authenticate: async (email) => {
    try {
      const response = await fetch("/api/passkey/authentication-begin", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email }),
      });
      if (!response.ok) {
        app.showError(await response.text(), false);
        return;
      }
      const options = await response.json();

      const assertionResponse = await SimpleWebAuthnBrowser.startAuthentication(
        { optionsJSON: options.publicKey },
      );

      const verificationResponse = await fetch(
        "/api/passkey/authentication-end",
        {
          method: "POST",
          credentials: "same-origin",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(assertionResponse),
        },
      );

      if (!verificationResponse.ok) {
        app.showError(await verificationResponse.text(), false);
        return;
      }
      const serverResponse = await verificationResponse.json();
      if (serverResponse.success) {
        app.Store.jwt = serverResponse.jwt;
        app.Router.go("/account/");
      } else {
        app.showError("We couldn't authenticate you using a Passkey", false);
      }
    } catch (e) {
      console.log(e);
      app.showError("We couldn't authenticate you using a Passkey", false);
    }
  },
};
