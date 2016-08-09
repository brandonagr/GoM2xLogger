$fn=100;

difference() {
    
union() {
    
    cube([60,11,2]);
    
    translate([0,0,2]) {
      cube([1.5,11,2]);  
    }
    
    translate([0,0,4]) {
       cube([2.7,11,1]);
    }
    
    translate([40,0,2]) {
      cube([1.5,11,2]);  
    }
    
    translate([40,0,4]) {
       cube([2.7,11,1]);
    }
}

    translate([55,11/2,0]){
    cylinder(h=10, r=4.1/2);
    }
}