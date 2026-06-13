;; Container for go-based boot.dev projects

(use-modules (gnu)
	     (gnosis packages learning-tools)
	     (gnu packages golang)
	     (gnu packages golang-xyz)
	     (gnu packages golang-build)
	     (gnu services networking)
             (gnu services containers)
             (gnu services databases)
	     (gnu packages databases)
             (gnu packages golang)
	     (gnu services ssh)
             (gnu packages version-control) ; Useful for git inside the container
             (gnu packages bash))

(operating-system
  (host-name "go-postgres-dev")
  (timezone "Etc/UTC")
  (locale "en_US.utf8")

  (users (cons (user-account
		(name "dev")
		(group "users")
		(supplementary-groups '("wheel")))
	       %base-user-accounts))
  
  ;; 1. Bootloader is required by the OS structure, 
  ;; but the container mechanism bypasses it.
  (bootloader (bootloader-configuration
               (bootloader grub-bootloader)
               (targets '("/dev/sdX"))))

  ;; 2. Set up a basic file system structure for the container
  (file-systems (cons (file-system
                        (device (file-system-label "container_root"))
                        (mount-point "/")
                        (type "ext4"))
                      %base-file-systems))

  ;; 3. Globally installed packages inside the container
  (packages (append (list go 
			  git 
			  bash-minimal 
			  postgresql 
			  bootdev-cli
			  go-github-com-lib-pq
			  go-github-com-google-uuid)
                    %base-packages))

  ;; 4. Services (This is where the magic happens)
  (services
   (append (list
	    (service openssh-service-type)
            (service postgresql-service-type
                     (postgresql-configuration
                      ;; Optional: permit connections from local container users without a password
		      (postgresql postgresql)
                      (config-file
                       (postgresql-config-file
                        (hba-file
                         (plain-file "pg_hba.conf"
                                     "local all all trust\nhost all all 127.0.0.1/32 trust")))))))
           %base-services)))
