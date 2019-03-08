Pod::Spec.new do |spec|
  spec.name         = 'Gman'
  spec.version      = '{{.Version}}'
  spec.license      = { :type => 'GNU Lesser General Public License, Version 3.0' }
  spec.homepage     = 'https://github.com/MATRIX/go-MATRIX'
  spec.authors      = { {{range .Contributors}}
		'{{.Name}}' => '{{.Email}}',{{end}}
	}
  spec.summary      = 'iOS MATRIX Client'
  spec.source       = { :git => 'https://github.com/MATRIX/go-MATRIX.git', :commit => '{{.Commit}}' }

	spec.platform = :ios
  spec.ios.deployment_target  = '9.0'
	spec.ios.vendored_frameworks = 'Frameworks/Gman.framework'

	spec.prepare_command = <<-CMD
    curl https://gmanstore.blob.core.windows.net/builds/{{.Archive}}.tar.gz | tar -xvz
    mkdir Frameworks
    mv {{.Archive}}/Gman.framework Frameworks
    rm -rf {{.Archive}}
  CMD
end
