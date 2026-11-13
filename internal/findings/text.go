package findings

import "github.com/XexusCode/hardenme/internal/check"

// Text explains one kind of finding in plain language.
type Text struct {
	// Title is a question: these are things to recognise, not verdicts.
	Title string
	Why   string
	// Do says what to do if the person does not recognise it.
	Do string
}

// TextFor returns the explanation of a kind, falling back to English.
func TextFor(kind string, lang check.Lang) Text {
	if t, ok := texts[lang][kind]; ok {
		return t
	}
	return texts[check.EN][kind]
}

var texts = map[check.Lang]map[string]Text{
	check.EN: {
		RemoteTool: {
			Title: "Did you install these remote access apps?",
			Why:   "These apps let another person see your screen or control your Mac from somewhere else. They are normal if you or your company's support installed them. They are also what phone scammers ask people to install, and what someone uses to watch another person.",
			Do:    "If you do not use it, quit it and drag it from Applications to the Trash. If someone told you to install it over the phone, change your passwords from another device and call your bank.",
		},
		StartupUnsigned: {
			Title: "Do you recognise these programs that start by themselves?",
			Why:   "They start every time you turn on the Mac, and they are not signed by Apple or by an identified developer, so nobody vouches for where they come from. Tools made by hobbyists or by yourself look like this. So does malware.",
			Do:    "Search for the name. If you cannot tell what it is, ask someone you trust before deleting it, and run a check with a reputable anti-malware tool.",
		},
		StartupScript: {
			Title: "Do you recognise these scripts that start by themselves?",
			Why:   "They run a script every time you turn on the Mac or sign in. Some apps and developer tools do this. It is also a common way for unwanted software to keep coming back.",
			Do:    "If you do not develop software and do not recognise the name, ask someone you trust to look at it with you.",
		},
		Profile: {
			Title: "Did you or your company install these configuration profiles?",
			Why:   "A profile can change security settings, add certificates that let someone read your encrypted traffic, or send your connection through another server. Companies and schools use them. A profile you did not install is a serious warning sign.",
			Do:    "Open System Settings > General > Device Management, look at each profile and remove the ones you do not recognise.",
		},
		Managed: {
			Title: "Is this Mac managed by your company or school?",
			Why:   "This Mac is enrolled in remote management. Whoever manages it can install apps, change settings and, in some cases, lock or erase it. That is expected on a work or school Mac, and not on a personal one.",
			Do:    "If it is your own Mac and nobody should manage it, open System Settings > General > Device Management to see who does, and ask whoever sold or gave you the Mac.",
		},
		ExtraAdmin: {
			Title: "Do you know these administrator accounts?",
			Why:   "Besides you, these accounts can install software and change any setting on this Mac. A family member or your company's support may have one. An account you do not know should not be there.",
			Do:    "Open System Settings > Users & Groups. Remove the accounts you do not recognise, or turn off \"Allow this user to administer this computer\" for them.",
		},
		SSHKeys: {
			Title: "Did you add these keys to sign in from another computer?",
			Why:   "Whoever holds one of these keys can sign in to your account on this Mac over the network without your password, if Remote Login is on. Developers add them on purpose. Nobody else should have one.",
			Do:    "If you do not use Remote Login, turn it off in System Settings > General > Sharing. To remove a key, delete its line from the file .ssh/authorized_keys in your home folder.",
		},
		Proxy: {
			Title: "Did you set up this proxy?",
			Why:   "Your Mac sends its web traffic through another server. Companies, some VPNs and ad blockers do this. A proxy you did not set up can see, and sometimes change, the sites you visit.",
			Do:    "Open System Settings > Network, choose your connection, then Details > Proxies, and turn off what you do not recognise.",
		},
		Hosts: {
			Title: "Did you add these redirections?",
			Why:   "A file on your Mac sends these site names to an address chosen by whoever edited it. Developers and ad blockers use it. Malware uses it to send you to a fake copy of a bank or an email site.",
			Do:    "If a name here is your bank, your email or a shop, do not sign in on this Mac until someone you trust has looked at the file /etc/hosts.",
		},
	},
	check.ES: {
		RemoteTool: {
			Title: "¿Instalaste tú estas apps de acceso remoto?",
			Why:   "Estas apps dejan que otra persona vea tu pantalla o maneje tu Mac desde otro sitio. Son normales si las instalaste tú o el soporte de tu empresa. También son las que piden instalar los timadores por teléfono, y las que usa alguien para vigilar a otra persona.",
			Do:    "Si no la usas, ciérrala y arrástrala de Aplicaciones a la Papelera. Si te la hizo instalar alguien por teléfono, cambia tus contraseñas desde otro dispositivo y llama a tu banco.",
		},
		StartupUnsigned: {
			Title: "¿Reconoces estos programas que arrancan solos?",
			Why:   "Arrancan cada vez que enciendes el Mac y no están firmados por Apple ni por un desarrollador identificado, así que nadie responde de dónde vienen. Así se ven las herramientas hechas por aficionados o por ti mismo. También el software malicioso.",
			Do:    "Busca el nombre en internet. Si no consigues saber qué es, pregunta a alguien de confianza antes de borrarlo y pasa una herramienta antimalware conocida.",
		},
		StartupScript: {
			Title: "¿Reconoces estos scripts que arrancan solos?",
			Why:   "Ejecutan un script cada vez que enciendes el Mac o inicias sesión. Algunas apps y herramientas de programación lo hacen. También es una forma habitual de que un programa no deseado vuelva una y otra vez.",
			Do:    "Si no programas y no reconoces el nombre, pide a alguien de confianza que lo mire contigo.",
		},
		Profile: {
			Title: "¿Instalaste tú o tu empresa estos perfiles de configuración?",
			Why:   "Un perfil puede cambiar ajustes de seguridad, añadir certificados que permiten leer tu tráfico cifrado o enviar tu conexión por otro servidor. Los usan empresas y colegios. Un perfil que no instalaste es una señal de alarma seria.",
			Do:    "Abre Ajustes del Sistema > General > Gestión de dispositivos, mira cada perfil y quita los que no reconozcas.",
		},
		Managed: {
			Title: "¿Este Mac lo administra tu empresa o tu centro de estudios?",
			Why:   "Este Mac está inscrito en una gestión remota. Quien lo administra puede instalar apps, cambiar ajustes y, en algunos casos, bloquearlo o borrarlo. Es lo esperable en un Mac del trabajo o del colegio, y no en uno personal.",
			Do:    "Si el Mac es tuyo y nadie debería administrarlo, abre Ajustes del Sistema > General > Gestión de dispositivos para ver quién lo hace, y pregunta a quien te lo vendió o te lo dio.",
		},
		ExtraAdmin: {
			Title: "¿Conoces estas cuentas de administrador?",
			Why:   "Además de ti, estas cuentas pueden instalar programas y cambiar cualquier ajuste de este Mac. Puede tener una un familiar o el soporte de tu empresa. Una cuenta que no conoces no debería estar ahí.",
			Do:    "Abre Ajustes del Sistema > Usuarios y grupos. Elimina las cuentas que no reconozcas, o desactívales \"Permitir que este usuario administre este ordenador\".",
		},
		SSHKeys: {
			Title: "¿Añadiste tú estas claves para entrar desde otro ordenador?",
			Why:   "Quien tenga una de estas claves puede entrar en tu cuenta de este Mac por la red sin tu contraseña, si el inicio de sesión remoto está activado. Los programadores las añaden a propósito. Nadie más debería tener una.",
			Do:    "Si no usas el inicio de sesión remoto, desactívalo en Ajustes del Sistema > General > Compartir. Para quitar una clave, borra su línea del archivo .ssh/authorized_keys de tu carpeta de usuario.",
		},
		Proxy: {
			Title: "¿Configuraste tú este proxy?",
			Why:   "Tu Mac envía su tráfico web a través de otro servidor. Lo hacen las empresas, algunas VPN y los bloqueadores de anuncios. Un proxy que no configuraste puede ver, y a veces cambiar, las webs que visitas.",
			Do:    "Abre Ajustes del Sistema > Red, elige tu conexión, luego Detalles > Proxies, y desactiva lo que no reconozcas.",
		},
		Hosts: {
			Title: "¿Añadiste tú estas redirecciones?",
			Why:   "Un archivo de tu Mac manda estos nombres de webs a la dirección que eligió quien lo editó. Lo usan programadores y bloqueadores de anuncios. El software malicioso lo usa para llevarte a una copia falsa de un banco o de un correo.",
			Do:    "Si alguno de estos nombres es tu banco, tu correo o una tienda, no inicies sesión en este Mac hasta que alguien de confianza haya mirado el archivo /etc/hosts.",
		},
	},
}
