package ru.tkonf.sign;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.autoconfigure.jdbc.DataSourceAutoConfiguration;
import org.springframework.boot.autoconfigure.orm.jpa.HibernateJpaAutoConfiguration;
import org.springframework.boot.autoconfigure.security.servlet.SecurityAutoConfiguration;
import org.springframework.context.annotation.Lazy;

/// Отключение ненужных автоконфигураций
@SpringBootApplication(
		scanBasePackages = "ru.tkonf.sign",
		exclude = {
		DataSourceAutoConfiguration.class,
		HibernateJpaAutoConfiguration.class,
		SecurityAutoConfiguration.class
},proxyBeanMethods = false)
@Lazy
public class TkonfAppSpringApplication {

	public static void main(String[] args) {
		SpringApplication.run(TkonfAppSpringApplication.class, args);
	}

}
