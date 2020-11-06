Pod::Spec.new do |spec|
  spec.name         = 'Gdch'
  spec.version      = '{{.Version}}'
  spec.license      = { :type => 'GNU Lesser General Public License, Version 3.0' }
  spec.homepage     = 'https://github.com/deltachaineum/go-deltachaineum'
  spec.authors      = { {{range .Contributors}}
		'{{.Name}}' => '{{.Email}}',{{end}}
	}
  spec.summary      = 'iOS Deltachain Client'
  spec.source       = { :git => 'https://github.com/deltachaineum/go-deltachaineum.git', :commit => '{{.Commit}}' }

	spec.platform = :ios
  spec.ios.deployment_target  = '9.0'
	spec.ios.vendored_frameworks = 'Frameworks/Gdch.framework'

	spec.prepare_command = <<-CMD
    curl https://gdchstore.blob.core.windows.net/builds/{{.Archive}}.tar.gz | tar -xvz
    mkdir Frameworks
    mv {{.Archive}}/Gdch.framework Frameworks
    rm -rf {{.Archive}}
  CMD
end
