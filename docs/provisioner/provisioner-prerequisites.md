intro: co to jest? komponent compassowy odpowiadający za provisioning i instalację klastrów z Kymą (Kyma Runtimes) - w tej chwili mozemy to zrobic w GCP lub z pomocą Gardenera w Azurze, GCP lub AWS (Amaozon Web Services). 

		Provisioning mozna podzielic na: 

			1) na GCP 

			Wszystkie te sposoby wymagają wcześniej pewnej konfiguracji. Jeśli chcemy postawić klaster na GCP, potrzebujemy mieć wcześniej utworzony projekt na GCP i potrzebujemy mieć Service Account z uprawnieniami, które dostałam na Slacku od Szymona.

		Teraz trzeba to bardzo precyzyjnie opisać, bo chodzi o to, że pobieram klucz dla Service Account'a. 
		(- Secret zawierający dane do Twojego Service Accounta. Żeby to zorbić, musisz utworzyć Service Account, nadać mu role, wygenerować i pobrać klucz w formacie json i stworzyć z niego Secret.)
		
		( Service Account z nadanymi odpowiednimi rolami z wygenerowanym kluczem, który musi się nazywać "credentials", )

			tl;dr: PREREQ do GCP: Istotne są 3 rzeczy: ten Secret musi mieć klucz "credentials", wartość musi być zakodowana base64, ten secret musi być w namespacesie compass-system.


			2) na Gardenerze
				- na GCP
				- na Azure
				- na AWS


			W tej chwili, żeby dostać się do Provisionera trzeba albo wykonać call z wewnątrz klastra na którym jest Provisioner (z jakiegoś innego poda), albo zrobić port-forward. (|Mogę wejść do jakiegoś poda, otworzyć shella i zrobić `kcl exec {POLECENIE_KTORE_ZOSTANIE_WYKONANE_WEWNATRZ_PODA}` jeśli mam curla.)

			`kubectl -n compass-system port-forward svc/compass-provisioner 3000:3000`
		

			```
			# Write your query or mutation here
			mutation { provisionRuntime(id:"309051b6-0bac-44c8-8bae-3fc59c12bb5c" config: {
			  clusterConfig: {
			    gcpConfig: {
			      name: "maja-test-gcp" ### nazwa klastra
			      projectName: "sap-kyma-framefrog-dev" ### USUNAC! wasy-> nazwa projektu
			      kubernetesVersion: "1.13"
			      bootDiskSizeGB: 30
			      numberOfNodes: 1
			      machineType: "n1-standard-4"
			      region: "europe-west3-a"
			    }
			  }
			  # instalacja Kymy jeszcze nie jest wspierana
			  kymaConfig: {
			    version: "1.5"
			    modules: Backup
			  }
			  credentials: {
			    secretName: "maja-gcp-secret" ### USUNAC -> nazwa secretu
			  }
			}
			) }
			```

			Runtime Operation Status:

			Operacje provisioningu i deprovisioningu są wykonywane asynchronicznie. 