# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure(2) do |config|
	config.vm.box = "chef/centos-6.6"
	config.vm.network "forwarded_port", guest: 8080, host: 8080
	config.vm.provider "virtualbox" do |vb|
		vb.memory = 512
		vb.cpus = 1
	end
	config.vm.provision "ansible" do |ansible|
		ansible.playbook = "provisioning/main.yml"
	end
end
